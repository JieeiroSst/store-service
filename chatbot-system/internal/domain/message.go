package domain

import "time"

type MessageType string

const (
	MessageTypeUser MessageType = "user"
	MessageTypeAI   MessageType = "ai"
)

type Message struct {
	ID             int64       `json:"id"`
	ConversationID int64       `json:"conversation_id"`
	SenderID       int64       `json:"sender_id"`
	Content        string      `json:"content"`
	MessageType    MessageType `json:"message_type"`
	AIModel        *string     `json:"ai_model,omitempty"` // claude, deepseek, etc.
	CreatedAt      time.Time   `json:"created_at"`
}

type Conversation struct {
	ID           int64     `json:"id"`
	User1ID      int64     `json:"user1_id"` // Manager/Advisor
	User2ID      int64     `json:"user2_id"` // User or can be AI conversation
	IsAIChat     bool      `json:"is_ai_chat"`
	ActiveAIModel *string  `json:"active_ai_model,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
