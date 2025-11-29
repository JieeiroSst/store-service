package repository

import (
	"context"
	"fmt"

	"chatbot-system/internal/core/domain"
	"gorm.io/gorm"
)

type conversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB) *conversationRepository {
	return &conversationRepository{db: db}
}

// CreateConversation creates a new conversation
func (r *conversationRepository) CreateConversation(ctx context.Context, conversation *domain.Conversation) error {
	result := r.db.WithContext(ctx).Create(conversation)
	if result.Error != nil {
		return fmt.Errorf("failed to create conversation: %w", result.Error)
	}
	return nil
}

// GetConversationByID retrieves a conversation by ID
func (r *conversationRepository) GetConversationByID(ctx context.Context, id uint) (*domain.Conversation, error) {
	var conversation domain.Conversation
	result := r.db.WithContext(ctx).First(&conversation, id)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", result.Error)
	}
	
	return &conversation, nil
}

// GetConversationsByUserID retrieves all conversations for a user
func (r *conversationRepository) GetConversationsByUserID(ctx context.Context, userID string) ([]domain.Conversation, error) {
	var conversations []domain.Conversation
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&conversations)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", result.Error)
	}
	
	return conversations, nil
}

// UpdateConversation updates an existing conversation
func (r *conversationRepository) UpdateConversation(ctx context.Context, conversation *domain.Conversation) error {
	result := r.db.WithContext(ctx).Save(conversation)
	if result.Error != nil {
		return fmt.Errorf("failed to update conversation: %w", result.Error)
	}
	return nil
}
