package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"aiagentsystem/internal/chat"
)

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type wsHandlers struct {
	service *chat.Service
}

type wsMessage struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Answer  string `json:"answer,omitempty"`
	Source  string `json:"source,omitempty"`
	Message string `json:"message,omitempty"`
}

type safeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *safeConn) writeJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
	return c.conn.WriteJSON(v)
}

func (c *safeConn) ping() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(wsWriteWait))
}

func (h *wsHandlers) Chat(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)

	rawConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already wrote an HTTP error response.
		return
	}
	defer rawConn.Close()
	conn := &safeConn{conn: rawConn}

	_ = rawConn.SetReadDeadline(time.Now().Add(wsPongWait))
	rawConn.SetPongHandler(func(string) error {
		return rawConn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	stopPing := make(chan struct{})
	defer close(stopPing)
	go func() {
		ticker := time.NewTicker(wsPingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.ping(); err != nil {
					return
				}
			case <-stopPing:
				return
			}
		}
	}()

	for {
		var msg ChatRequest
		if err := rawConn.ReadJSON(&msg); err != nil {
			// Client closed the connection, sent a malformed frame, or the
			// read deadline (no pong received) was exceeded — either way,
			// the connection is done.
			return
		}
		if msg.Question == "" {
			_ = conn.writeJSON(wsMessage{Type: "error", Message: "question must not be empty"})
			continue
		}

		result, err := h.service.AnswerQuestionStream(r.Context(), chat.AnswerRequest{
			Messages: []chat.Message{{Role: "user", Content: msg.Question}},
			UserID:   userID,
		}, func(delta string) {
			_ = conn.writeJSON(wsMessage{Type: "delta", Text: delta})
		})
		if err != nil {
			_, _, message := errorTypeAndStatus(err)
			_ = conn.writeJSON(wsMessage{Type: "error", Message: message})
			continue
		}

		_ = conn.writeJSON(wsMessage{Type: "done", Answer: result.Text, Source: string(result.Source)})
	}
}
