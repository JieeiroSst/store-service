package repository

import (
	"context"
	"fmt"

	"chatbot-system/internal/core/domain"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) *messageRepository {
	return &messageRepository{db: db}
}

// CreateMessage saves a new message to the database
func (r *messageRepository) CreateMessage(ctx context.Context, message *domain.Message) error {
	result := r.db.WithContext(ctx).Create(message)
	if result.Error != nil {
		return fmt.Errorf("failed to create message: %w", result.Error)
	}
	return nil
}

// GetMessagesByConversationID retrieves all messages for a conversation
func (r *messageRepository) GetMessagesByConversationID(ctx context.Context, conversationID uint) ([]domain.Message, error) {
	var messages []domain.Message
	result := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get messages: %w", result.Error)
	}
	
	return messages, nil
}

// GetConversationHistory retrieves the last N messages for a conversation
func (r *messageRepository) GetConversationHistory(ctx context.Context, conversationID uint, limit int) ([]domain.Message, error) {
	var messages []domain.Message
	result := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", result.Error)
	}
	
	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, nil
}
