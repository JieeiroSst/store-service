package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents the chat message structure
type Message struct {
	Type        string `json:"type"`
	MessageType string `json:"messageType"`
	SenderID    string `json:"sender_id"`
	Data        struct {
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LlamaRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type StreamResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// Client represents a connected WebSocket client
type Client struct {
	Conn     *websocket.Conn
	SendChan chan Message
	UserID   string
}

// ChatServer manages WebSocket connections and broadcasts
type ChatServer struct {
	Clients    map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.RWMutex
}

// NewChatServer initializes a new chat server
func NewChatServer() *ChatServer {
	return &ChatServer{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message, 1000),
		Register:   make(chan *Client, 100),
		Unregister: make(chan *Client, 100),
	}
}

// Run manages client connections and message broadcasting
func (server *ChatServer) Run() {
	for {
		select {
		case client := <-server.Register:
			server.Mutex.Lock()
			server.Clients[client] = true
			server.Mutex.Unlock()

		case client := <-server.Unregister:
			server.Mutex.Lock()
			if _, ok := server.Clients[client]; ok {
				delete(server.Clients, client)
				close(client.SendChan)
			}
			server.Mutex.Unlock()

		case message := <-server.Broadcast:
			server.Mutex.RLock()
			for client := range server.Clients {
				select {
				case client.SendChan <- message:
				default:
					close(client.SendChan)
					delete(server.Clients, client)
				}
			}
			server.Mutex.RUnlock()
		}
	}
}

// HandleWebSocket manages WebSocket connection and messaging
func (server *ChatServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("user_id")
	client := &Client{
		Conn:     conn,
		SendChan: make(chan Message, 256),
		UserID:   userID,
	}

	server.Register <- client

	go server.writePump(client)
	server.readPump(client)
}

// writePump sends messages to the client
func (server *ChatServer) writePump(client *Client) {
	defer func() {
		client.Conn.Close()
		server.Unregister <- client
	}()

	for {
		select {
		case message, ok := <-client.SendChan:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := client.Conn.WriteJSON(message)
			if err != nil {
				return
			}
		}
	}
}

// readPump receives messages from the client
func (server *ChatServer) readPump(client *Client) {
	defer func() {
		server.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, payload, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var message Message
		if err := json.Unmarshal(payload, &message); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		message.Timestamp = time.Now()
		server.processMessage(message)
	}
}

// processMessage handles different message types
func (server *ChatServer) processMessage(message Message) {
	switch message.Type {
	case "bot":
		go server.handleBotMessage(message)
	case "consultant":
		go server.handleConsultantMessage(message)
	default:
		log.Println("Unknown message type:", message.Type)
	}
}

// handleBotMessage processes bot responses
func (server *ChatServer) handleBotMessage(message Message) {
	// Integration with Ollama API for bot responses
	botResponse, err := fetchOllamaResponse(message.Data.Content)
	if err != nil {
		log.Println("Bot response error:", err)
		return
	}

	botMessage := Message{
		Type:        "bot",
		MessageType: "text",
		SenderID:    "system",
		Data: struct {
			URL     string `json:"url"`
			Content string `json:"content"`
		}{
			Content: botResponse,
		},
		Timestamp: time.Now(),
	}

	aa, _ := json.Marshal(&botMessage)
	fmt.Println("=======", string(aa))

	server.Broadcast <- botMessage
}

// handleConsultantMessage routes messages to consultant API
func (server *ChatServer) handleConsultantMessage(message Message) {
	// Implement consultant API routing logic
	consultantMessage := message
	server.Broadcast <- consultantMessage
}

// fetchOllamaResponse handles chat with Ollama API
func fetchOllamaResponse(message string) (string, error) {
	req := LlamaRequest{
		Model: "Tuanpham/t-visstar-7b",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: message,
			},
		},
		Stream: true,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", "http://localhost:11434/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var fullResponse string

	for {
		var streamResp StreamResponse
		if err := decoder.Decode(&streamResp); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		fullResponse += streamResp.Message.Content

		if streamResp.Done {
			break
		}
	}

	return fullResponse, nil
}

func main() {
	chatServer := NewChatServer()
	go chatServer.Run()

	http.HandleFunc("/ws", chatServer.HandleWebSocket)
	log.Fatal(http.ListenAndServe(":8083", nil))
}
