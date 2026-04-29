package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"ollama-gateway/internal/ollama"
	"ollama-gateway/internal/preset"
	"ollama-gateway/internal/search"
)

const (
	searchThreshold = 0.5

	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Config struct {
	Addr        string
	PresetsDir  string
	OllamaURL   string
	OllamaModel string
}

type ClientMessage struct {
	Text   string   `json:"text"`
	Images []string `json:"images,omitempty"` // base64 or file paths
}

// ServerMessage is what we send back to the client
type ServerMessage struct {
	Type    string `json:"type"`    // "token" | "done" | "error" | "preset"
	Content string `json:"content"`
	Source  string `json:"source"` // "preset" | "ollama"
}

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins in dev
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Hub manages all connected clients
type Hub struct {
	cfg         Config
	searchEngine *search.Engine
	ollamaClient *ollama.Client
}

func newHub(cfg Config) (*Hub, error) {
	loader := preset.NewLoader(cfg.PresetsDir)
	items, err := loader.Load()
	if err != nil {
		return nil, err
	}
	log.Printf("📚 Loaded %d preset Q&A items", len(items))

	// Build search index
	engine := search.NewEngine()
	engine.Build(items)
	log.Printf("🔍 %s", engine.DebugIndex())

	return &Hub{
		cfg:          cfg,
		searchEngine: engine,
		ollamaClient: ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel),
	}, nil
}

func Run(cfg Config) error {
	hub, err := newHub(cfg)
	if err != nil {
		return err
	}

	http.HandleFunc("/ws", hub.handleWS)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	http.HandleFunc("/", serveTestPage)

	return http.ListenAndServe(cfg.Addr, nil)
}

func (h *Hub) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("🔗 Client connected: %s", r.RemoteAddr)

	var (
		history []ollama.Message
		mu      sync.Mutex
	)

	// Ping to keep connection alive
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS error: %v", err)
			}
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			h.sendError(conn, "Invalid message format")
			continue
		}

		if msg.Text == "" {
			h.sendError(conn, "Empty message")
			continue
		}

		log.Printf("📨 [%s] Query: %q (images: %d)", r.RemoteAddr, msg.Text, len(msg.Images))

		// ── Route decision ────────────────────────────────────────────────
		// 1. Try full-text search in presets
		result := h.searchEngine.Search(msg.Text, searchThreshold)

		if result != nil {
			// Case 1: Preset answer found
			log.Printf("✅ Preset match (score=%.2f): %s", result.Score, result.QA.ID)

			h.sendMsg(conn, ServerMessage{
				Type:    "preset",
				Content: result.QA.Answer,
				Source:  "preset",
			})

			// Optionally send image if preset has one
			if result.QA.ImageURL != "" {
				h.sendMsg(conn, ServerMessage{
					Type:    "image",
					Content: result.QA.ImageURL,
					Source:  "preset",
				})
			}

			h.sendMsg(conn, ServerMessage{Type: "done", Source: "preset"})

			// Add to history for context continuity
			mu.Lock()
			history = append(history,
				ollama.Message{Role: "user", Content: msg.Text},
				ollama.Message{Role: "assistant", Content: result.QA.Answer},
			)
			mu.Unlock()

		} else {
			// Case 2: No preset match → send to Ollama (which can call external APIs)
			log.Printf("🤖 Routing to Ollama model: %s", h.cfg.OllamaModel)

			mu.Lock()
			currentHistory := make([]ollama.Message, len(history))
			copy(currentHistory, history)
			mu.Unlock()

			var fullResponse strings.Builder

			err := h.ollamaClient.StreamChat(
				currentHistory,
				msg.Text,
				msg.Images,
				func(chunk ollama.StreamChunk) {
					switch chunk.Type {
					case "token":
						fullResponse.WriteString(chunk.Content)
						h.sendMsg(conn, ServerMessage{
							Type:    "token",
							Content: chunk.Content,
							Source:  "ollama",
						})
					case "done":
						h.sendMsg(conn, ServerMessage{Type: "done", Source: "ollama"})
					case "error":
						h.sendError(conn, chunk.Content)
					}
				},
			)

			if err != nil {
				log.Printf("❌ Ollama error: %v", err)
				h.sendError(conn, "AI service error: "+err.Error())
				continue
			}

			// Update history
			mu.Lock()
			history = append(history,
				ollama.Message{Role: "user", Content: msg.Text},
				ollama.Message{Role: "assistant", Content: fullResponse.String()},
			)
			// Keep history bounded (last 20 turns)
			if len(history) > 40 {
				history = history[len(history)-40:]
			}
			mu.Unlock()
		}
	}

	log.Printf("🔌 Client disconnected: %s", r.RemoteAddr)
}

func (h *Hub) sendMsg(conn *websocket.Conn, msg ServerMessage) {
	data, _ := json.Marshal(msg)
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	conn.WriteMessage(websocket.TextMessage, data)
}

func (h *Hub) sendError(conn *websocket.Conn, errMsg string) {
	h.sendMsg(conn, ServerMessage{Type: "error", Content: errMsg})
}

func serveTestPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(testPageHTML))
}

const testPageHTML = `<!DOCTYPE html>
<html lang="vi">
<head>
<meta charset="UTF-8">
<title>Ollama Gateway Test</title>
<style>
  * { box-sizing: border-box; }
  body { font-family: monospace; max-width: 800px; margin: 40px auto; padding: 0 20px; background: #0d1117; color: #c9d1d9; }
  h1 { color: #58a6ff; }
  #chat { height: 400px; overflow-y: auto; border: 1px solid #30363d; padding: 12px; border-radius: 8px; margin-bottom: 12px; background: #161b22; }
  .msg { margin: 8px 0; padding: 8px 12px; border-radius: 6px; }
  .user { background: #1f6feb22; border-left: 3px solid #1f6feb; }
  .assistant { background: #238636222; border-left: 3px solid #238636; }
  .preset { background: #9e6a0322; border-left: 3px solid #e3b341; }
  .error { background: #da363322; border-left: 3px solid #da3633; }
  .source-badge { font-size: 11px; opacity: 0.6; }
  #form { display: flex; gap: 8px; }
  #input { flex: 1; padding: 10px; border: 1px solid #30363d; border-radius: 6px; background: #161b22; color: #c9d1d9; font-family: monospace; }
  button { padding: 10px 20px; background: #238636; color: white; border: none; border-radius: 6px; cursor: pointer; }
  button:hover { background: #2ea043; }
  #status { margin-top: 8px; font-size: 12px; color: #8b949e; }
</style>
</head>
<body>
<h1>🤖 Ollama Gateway</h1>
<div id="chat"></div>
<div id="form">
  <input id="input" type="text" placeholder="Nhập câu hỏi..." autofocus />
  <button onclick="send()">Gửi</button>
</div>
<div id="status">Connecting...</div>
<script>
const chat = document.getElementById('chat');
const input = document.getElementById('input');
const status = document.getElementById('status');
let ws, currentDiv = null;

function connect() {
  ws = new WebSocket('ws://' + location.host + '/ws');
  ws.onopen = () => { status.textContent = '✅ Connected'; };
  ws.onclose = () => { status.textContent = '❌ Disconnected. Reconnecting...'; setTimeout(connect, 2000); };
  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if (msg.type === 'preset') {
      addMsg(msg.content, 'preset', '⚡ preset');
      currentDiv = null;
    } else if (msg.type === 'token') {
      if (!currentDiv) { currentDiv = addMsg('', 'assistant', '🤖 ollama'); }
      currentDiv.querySelector('.text').textContent += msg.content;
      chat.scrollTop = chat.scrollHeight;
    } else if (msg.type === 'done') {
      currentDiv = null;
    } else if (msg.type === 'error') {
      addMsg(msg.content, 'error', '❌ error');
      currentDiv = null;
    }
  };
}

function addMsg(text, cls, source) {
  const d = document.createElement('div');
  d.className = 'msg ' + cls;
  d.innerHTML = '<span class="source-badge">' + source + '</span><br><span class="text">' + text + '</span>';
  chat.appendChild(d);
  chat.scrollTop = chat.scrollHeight;
  return d;
}

function send() {
  const text = input.value.trim();
  if (!text || ws.readyState !== 1) return;
  addMsg(text, 'user', '👤 you');
  ws.send(JSON.stringify({ text }));
  input.value = '';
}

input.addEventListener('keydown', e => { if (e.key === 'Enter') send(); });
connect();
</script>
</body>
</html>`


