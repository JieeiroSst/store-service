package domain

import (
	"time"
)

// Message represents a chat message
type Message struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ConversationID uint      `json:"conversation_id" gorm:"index"`
	Role           string    `json:"role"` // "user" or "assistant"
	Content        string    `json:"content" gorm:"type:text"`
	AIModel        string    `json:"ai_model"` // "claude", "deepseek"
	CreatedAt      time.Time `json:"created_at"`
}

// Conversation represents a chat conversation
type Conversation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index"`
	Title     string    `json:"title"`
	AIModel   string    `json:"ai_model"` // Current AI model being used
	Messages  []Message `json:"messages" gorm:"foreignKey:ConversationID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatRequest represents incoming chat request
type ChatRequest struct {
	ConversationID uint   `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Message        string `json:"message"`
	AIModel        string `json:"ai_model"` // "claude" or "deepseek"
}

// ChatResponse represents AI response
type ChatResponse struct {
	ConversationID uint      `json:"conversation_id"`
	MessageID      uint      `json:"message_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	AIModel        string    `json:"ai_model"`
	Timestamp      time.Time `json:"timestamp"`
}
