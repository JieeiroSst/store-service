package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"chatbot-system/internal/application"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure properly in production
	},
}

type Handler struct {
	hub         *Hub
	chatUseCase *application.ChatUseCase
}

func NewHandler(hub *Hub, chatUseCase *application.ChatUseCase) *Handler {
	return &Handler{
		hub:         hub,
		chatUseCase: chatUseCase,
	}
}

type gorillаConnection struct {
	*websocket.Conn
}

func (g *gorillаConnection) WriteMessage(messageType int, data []byte) error {
	return g.Conn.WriteMessage(messageType, data)
}

func (g *gorillаConnection) ReadMessage() (messageType int, p []byte, err error) {
	return g.Conn.ReadMessage()
}

func (g *gorillаConnection) Close() error {
	return g.Conn.Close()
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// In production, implement proper authentication
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:    h.hub,
		Conn:   &gorillаConnection{Conn: conn},
		Send:   make(chan []byte, 256),
		UserID: userID,
	}

	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump(h.createMessageHandler(client))
}

func (h *Handler) createMessageHandler(client *Client) func([]byte) error {
	return func(data []byte) error {
		var wsMsg WSMessage
		if err := json.Unmarshal(data, &wsMsg); err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}

		switch wsMsg.Type {
		case "send_message":
			return h.handleSendMessage(client, wsMsg.Payload)
		case "switch_ai_model":
			return h.handleSwitchAIModel(client, wsMsg.Payload)
		default:
			return fmt.Errorf("unknown message type: %s", wsMsg.Type)
		}
	}
}

type SendMessagePayload struct {
	RecipientID    *int64  `json:"recipient_id,omitempty"`
	ConversationID *int64  `json:"conversation_id,omitempty"`
	Content        string  `json:"content"`
	AIModel        *string `json:"ai_model,omitempty"`
}

type MessageResponse struct {
	Type         string `json:"type"`
	MessageID    int64  `json:"message_id"`
	SenderID     int64  `json:"sender_id"`
	Content      string `json:"content"`
	MessageType  string `json:"message_type"`
	AIModel      *string `json:"ai_model,omitempty"`
	Conversation int64  `json:"conversation_id"`
}

func (h *Handler) handleSendMessage(client *Client, payload json.RawMessage) error {
	var input SendMessagePayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Call use case
	output, err := h.chatUseCase.SendMessage(context.Background(), application.SendMessageInput{
		SenderID:       client.UserID,
		RecipientID:    input.RecipientID,
		ConversationID: input.ConversationID,
		Content:        input.Content,
		AIModel:        input.AIModel,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Send user message back to sender
	userMsgResp := MessageResponse{
		Type:         "message",
		MessageID:    output.Message.ID,
		SenderID:     output.Message.SenderID,
		Content:      output.Message.Content,
		MessageType:  string(output.Message.MessageType),
		Conversation: output.Conversation.ID,
	}

	if data, err := json.Marshal(userMsgResp); err == nil {
		client.Send <- data
	}

	// If there's a recipient (not AI chat), send to them
	if input.RecipientID != nil {
		if data, err := json.Marshal(userMsgResp); err == nil {
			h.hub.SendToUser(*input.RecipientID, data)
		}
	}

	// If AI responded, send AI message
	if output.AIResponse != nil {
		aiMsgResp := MessageResponse{
			Type:         "message",
			MessageID:    output.AIResponse.ID,
			SenderID:     0, // AI
			Content:      output.AIResponse.Content,
			MessageType:  string(output.AIResponse.MessageType),
			AIModel:      output.AIResponse.AIModel,
			Conversation: output.Conversation.ID,
		}

		if data, err := json.Marshal(aiMsgResp); err == nil {
			client.Send <- data
		}
	}

	return nil
}

type SwitchAIModelPayload struct {
	ConversationID int64  `json:"conversation_id"`
	NewModel       string `json:"new_model"`
}

func (h *Handler) handleSwitchAIModel(client *Client, payload json.RawMessage) error {
	var input SwitchAIModelPayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if err := h.chatUseCase.SwitchAIModel(context.Background(), client.UserID, input.ConversationID, input.NewModel); err != nil {
		return fmt.Errorf("failed to switch AI model: %w", err)
	}

	// Send confirmation
	response := map[string]interface{}{
		"type":            "ai_model_switched",
		"conversation_id": input.ConversationID,
		"new_model":       input.NewModel,
	}

	if data, err := json.Marshal(response); err == nil {
		client.Send <- data
	}

	return nil
}
