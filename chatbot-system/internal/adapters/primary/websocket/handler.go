package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"chatbot-system/internal/core/domain"
	"chatbot-system/internal/core/ports"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure this properly in production
	},
}

type Client struct {
	conn           *websocket.Conn
	send           chan []byte
	conversationID uint
	userID         string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.userID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

type WebSocketHandler struct {
	hub         *Hub
	chatService ports.ChatService
}

func NewWebSocketHandler(chatService ports.ChatService) *WebSocketHandler {
	hub := NewHub()
	go hub.Run()
	
	return &WebSocketHandler{
		hub:         hub,
		chatService: chatService,
	}
}

type WSMessage struct {
	Type           string `json:"type"` // "message", "history", "switch_model"
	ConversationID uint   `json:"conversation_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	Content        string `json:"content,omitempty"`
	AIModel        string `json:"ai_model,omitempty"`
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	userID := c.Query("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	h.hub.register <- client

	// Start goroutines for reading and writing
	go h.writePump(client)
	go h.readPump(client)
}

func (h *WebSocketHandler) readPump(client *Client) {
	defer func() {
		h.hub.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		switch wsMsg.Type {
		case "message":
			h.handleChatMessage(client, wsMsg)
		case "history":
			h.handleHistoryRequest(client, wsMsg)
		case "switch_model":
			h.handleSwitchModel(client, wsMsg)
		}
	}
}

func (h *WebSocketHandler) writePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) handleChatMessage(client *Client, wsMsg WSMessage) {
	ctx := context.Background()

	request := domain.ChatRequest{
		ConversationID: wsMsg.ConversationID,
		UserID:         client.userID,
		Message:        wsMsg.Content,
		AIModel:        wsMsg.AIModel,
	}

	response, err := h.chatService.ProcessMessage(ctx, request)
	if err != nil {
		h.sendError(client, err.Error())
		return
	}

	// Update client's conversation ID
	client.conversationID = response.ConversationID

	// Send response back to client
	responseMsg := map[string]interface{}{
		"type":            "response",
		"conversation_id": response.ConversationID,
		"message_id":      response.MessageID,
		"role":            response.Role,
		"content":         response.Content,
		"ai_model":        response.AIModel,
		"timestamp":       response.Timestamp,
	}

	jsonResponse, _ := json.Marshal(responseMsg)
	client.send <- jsonResponse
}

func (h *WebSocketHandler) handleHistoryRequest(client *Client, wsMsg WSMessage) {
	ctx := context.Background()

	messages, err := h.chatService.GetChatHistory(ctx, wsMsg.ConversationID)
	if err != nil {
		h.sendError(client, err.Error())
		return
	}

	responseMsg := map[string]interface{}{
		"type":     "history",
		"messages": messages,
	}

	jsonResponse, _ := json.Marshal(responseMsg)
	client.send <- jsonResponse
}

func (h *WebSocketHandler) handleSwitchModel(client *Client, wsMsg WSMessage) {
	ctx := context.Background()

	err := h.chatService.SwitchAIModel(ctx, wsMsg.ConversationID, wsMsg.AIModel)
	if err != nil {
		h.sendError(client, err.Error())
		return
	}

	responseMsg := map[string]interface{}{
		"type":      "model_switched",
		"ai_model":  wsMsg.AIModel,
		"message":   "AI model switched successfully",
	}

	jsonResponse, _ := json.Marshal(responseMsg)
	client.send <- jsonResponse
}

func (h *WebSocketHandler) sendError(client *Client, errorMsg string) {
	responseMsg := map[string]interface{}{
		"type":  "error",
		"error": errorMsg,
	}
	jsonResponse, _ := json.Marshal(responseMsg)
	client.send <- jsonResponse
}
