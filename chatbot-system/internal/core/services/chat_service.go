package services

import (
	"context"
	"fmt"
	"time"

	"chatbot-system/internal/core/domain"
	"chatbot-system/internal/core/ports"
)

type chatService struct {
	messageRepo      ports.MessageRepository
	conversationRepo ports.ConversationRepository
	aiFactory        *AIProviderFactory
}

// NewChatService creates a new chat service
func NewChatService(
	messageRepo ports.MessageRepository,
	conversationRepo ports.ConversationRepository,
	aiFactory *AIProviderFactory,
) ports.ChatService {
	return &chatService{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		aiFactory:        aiFactory,
	}
}

// ProcessMessage handles incoming chat messages
func (s *chatService) ProcessMessage(ctx context.Context, request domain.ChatRequest) (*domain.ChatResponse, error) {
	// Get or create conversation
	var conversation *domain.Conversation
	var err error

	if request.ConversationID == 0 {
		// Create new conversation
		conversation = &domain.Conversation{
			UserID:    request.UserID,
			Title:     "New Conversation",
			AIModel:   request.AIModel,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.conversationRepo.CreateConversation(ctx, conversation); err != nil {
			return nil, fmt.Errorf("failed to create conversation: %w", err)
		}
	} else {
		conversation, err = s.conversationRepo.GetConversationByID(ctx, request.ConversationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation: %w", err)
		}

		// Update AI model if changed
		if request.AIModel != "" && request.AIModel != conversation.AIModel {
			conversation.AIModel = request.AIModel
			conversation.UpdatedAt = time.Now()
			if err := s.conversationRepo.UpdateConversation(ctx, conversation); err != nil {
				return nil, fmt.Errorf("failed to update conversation: %w", err)
			}
		}
	}

	// Save user message
	userMessage := &domain.Message{
		ConversationID: conversation.ID,
		Role:           "user",
		Content:        request.Message,
		AIModel:        conversation.AIModel,
		CreatedAt:      time.Now(),
	}
	if err := s.messageRepo.CreateMessage(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get conversation history
	history, err := s.messageRepo.GetConversationHistory(ctx, conversation.ID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Get AI provider using Strategy Pattern
	aiProvider, err := s.aiFactory.GetProvider(conversation.AIModel)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	// Send message to AI
	aiResponse, err := aiProvider.SendMessage(ctx, history, request.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Save AI response
	assistantMessage := &domain.Message{
		ConversationID: conversation.ID,
		Role:           "assistant",
		Content:        aiResponse,
		AIModel:        conversation.AIModel,
		CreatedAt:      time.Now(),
	}
	if err := s.messageRepo.CreateMessage(ctx, assistantMessage); err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	return &domain.ChatResponse{
		ConversationID: conversation.ID,
		MessageID:      assistantMessage.ID,
		Role:           "assistant",
		Content:        aiResponse,
		AIModel:        conversation.AIModel,
		Timestamp:      assistantMessage.CreatedAt,
	}, nil
}

// GetChatHistory retrieves chat history for a conversation
func (s *chatService) GetChatHistory(ctx context.Context, conversationID uint) ([]domain.Message, error) {
	return s.messageRepo.GetMessagesByConversationID(ctx, conversationID)
}

// CreateConversation creates a new conversation
func (s *chatService) CreateConversation(ctx context.Context, userID, title, aiModel string) (*domain.Conversation, error) {
	conversation := &domain.Conversation{
		UserID:    userID,
		Title:     title,
		AIModel:   aiModel,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := s.conversationRepo.CreateConversation(ctx, conversation); err != nil {
		return nil, err
	}
	
	return conversation, nil
}

// SwitchAIModel switches the AI model for a conversation
func (s *chatService) SwitchAIModel(ctx context.Context, conversationID uint, newModel string) error {
	conversation, err := s.conversationRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return err
	}

	// Validate that the new model exists
	if _, err := s.aiFactory.GetProvider(newModel); err != nil {
		return fmt.Errorf("invalid AI model: %w", err)
	}

	conversation.AIModel = newModel
	conversation.UpdatedAt = time.Now()
	
	return s.conversationRepo.UpdateConversation(ctx, conversation)
}
