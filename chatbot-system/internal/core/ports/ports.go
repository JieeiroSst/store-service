package ports

import (
	"context"
	"chatbot-system/internal/core/domain"
)

// AIProvider defines the interface for AI service providers (Strategy Pattern)
type AIProvider interface {
	SendMessage(ctx context.Context, messages []domain.Message, userMessage string) (string, error)
	GetModelName() string
}

// MessageRepository defines the interface for message persistence
type MessageRepository interface {
	CreateMessage(ctx context.Context, message *domain.Message) error
	GetMessagesByConversationID(ctx context.Context, conversationID uint) ([]domain.Message, error)
	GetConversationHistory(ctx context.Context, conversationID uint, limit int) ([]domain.Message, error)
}

// ConversationRepository defines the interface for conversation persistence
type ConversationRepository interface {
	CreateConversation(ctx context.Context, conversation *domain.Conversation) error
	GetConversationByID(ctx context.Context, id uint) (*domain.Conversation, error)
	GetConversationsByUserID(ctx context.Context, userID string) ([]domain.Conversation, error)
	UpdateConversation(ctx context.Context, conversation *domain.Conversation) error
}

// ChatService defines the use case interface
type ChatService interface {
	ProcessMessage(ctx context.Context, request domain.ChatRequest) (*domain.ChatResponse, error)
	GetChatHistory(ctx context.Context, conversationID uint) ([]domain.Message, error)
	CreateConversation(ctx context.Context, userID, title, aiModel string) (*domain.Conversation, error)
	SwitchAIModel(ctx context.Context, conversationID uint, newModel string) error
}
