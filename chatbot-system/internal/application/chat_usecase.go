package application

import (
	"context"
	"fmt"

	"chatbot-system/internal/domain"
)

type ChatUseCase struct {
	userRepo         domain.UserRepository
	conversationRepo domain.ConversationRepository
	messageRepo      domain.MessageRepository
	aiFactory        domain.AIProviderFactory
}

func NewChatUseCase(
	userRepo domain.UserRepository,
	conversationRepo domain.ConversationRepository,
	messageRepo domain.MessageRepository,
	aiFactory domain.AIProviderFactory,
) *ChatUseCase {
	return &ChatUseCase{
		userRepo:         userRepo,
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		aiFactory:        aiFactory,
	}
}

type SendMessageInput struct {
	SenderID       int64
	RecipientID    *int64  // nil if AI chat
	ConversationID *int64  // nil if new conversation
	Content        string
	AIModel        *string // claude, deepseek, etc.
}

type SendMessageOutput struct {
	Message      *domain.Message
	AIResponse   *domain.Message
	Conversation *domain.Conversation
}

func (uc *ChatUseCase) SendMessage(ctx context.Context, input SendMessageInput) (*SendMessageOutput, error) {
	// Get sender
	sender, err := uc.userRepo.GetByID(ctx, input.SenderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender: %w", err)
	}

	var conversation *domain.Conversation

	// Check if this is an AI chat
	isAIChat := input.RecipientID == nil

	if input.ConversationID != nil {
		// Use existing conversation
		conversation, err = uc.conversationRepo.GetByID(ctx, *input.ConversationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation: %w", err)
		}

		// Verify user has access to this conversation
		if conversation.User1ID != sender.ID && conversation.User2ID != sender.ID {
			return nil, fmt.Errorf("user does not have access to this conversation")
		}
	} else {
		// Create new conversation
		conversation = &domain.Conversation{
			User1ID:       sender.ID,
			IsAIChat:      isAIChat,
			ActiveAIModel: input.AIModel,
		}

		if !isAIChat {
			// Human to human chat - verify permissions
			recipient, err := uc.userRepo.GetByID(ctx, *input.RecipientID)
			if err != nil {
				return nil, fmt.Errorf("failed to get recipient: %w", err)
			}

			if !sender.CanChatWith(recipient) {
				return nil, fmt.Errorf("sender does not have permission to chat with recipient")
			}

			conversation.User2ID = *input.RecipientID

			// Check if conversation already exists
			existingConv, err := uc.conversationRepo.GetByUsers(ctx, sender.ID, *input.RecipientID)
			if err != nil {
				return nil, fmt.Errorf("failed to check existing conversation: %w", err)
			}

			if existingConv != nil {
				conversation = existingConv
			} else {
				if err := uc.conversationRepo.Create(ctx, conversation); err != nil {
					return nil, fmt.Errorf("failed to create conversation: %w", err)
				}
			}
		} else {
			// AI chat
			conversation.User2ID = 0 // Special value for AI
			if err := uc.conversationRepo.Create(ctx, conversation); err != nil {
				return nil, fmt.Errorf("failed to create conversation: %w", err)
			}
		}
	}

	// Save user message
	userMessage := &domain.Message{
		ConversationID: conversation.ID,
		SenderID:       sender.ID,
		Content:        input.Content,
		MessageType:    domain.MessageTypeUser,
	}

	if err := uc.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	output := &SendMessageOutput{
		Message:      userMessage,
		Conversation: conversation,
	}

	// If AI chat, get AI response
	if conversation.IsAIChat {
		modelName := "claude" // default
		if input.AIModel != nil {
			modelName = *input.AIModel
		} else if conversation.ActiveAIModel != nil {
			modelName = *conversation.ActiveAIModel
		}

		// Get AI provider
		provider, err := uc.aiFactory.GetProvider(modelName)
		if err != nil {
			return nil, fmt.Errorf("failed to get AI provider: %w", err)
		}

		// Get conversation history
		historyPtrs, err := uc.messageRepo.GetConversationHistory(ctx, conversation.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation history: %w", err)
		}

		history := make([]domain.Message, len(historyPtrs))
		for i, msg := range historyPtrs {
			history[i] = *msg
		}

		// Get AI response
		aiResponse, err := provider.SendMessage(ctx, history, input.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to get AI response: %w", err)
		}

		// Save AI message
		aiMessage := &domain.Message{
			ConversationID: conversation.ID,
			SenderID:       0, // Special value for AI
			Content:        aiResponse,
			MessageType:    domain.MessageTypeAI,
			AIModel:        &modelName,
		}

		if err := uc.messageRepo.Create(ctx, aiMessage); err != nil {
			return nil, fmt.Errorf("failed to save AI message: %w", err)
		}

		output.AIResponse = aiMessage

		// Update conversation's active AI model
		conversation.ActiveAIModel = &modelName
		if err := uc.conversationRepo.Update(ctx, conversation); err != nil {
			return nil, fmt.Errorf("failed to update conversation: %w", err)
		}
	}

	return output, nil
}

func (uc *ChatUseCase) GetConversationHistory(ctx context.Context, userID, conversationID int64, limit, offset int) ([]*domain.Message, error) {
	// Get conversation
	conversation, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Verify user has access
	if conversation.User1ID != userID && conversation.User2ID != userID {
		return nil, fmt.Errorf("user does not have access to this conversation")
	}

	// Get messages
	messages, err := uc.messageRepo.GetConversationMessages(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

func (uc *ChatUseCase) GetUserConversations(ctx context.Context, userID int64) ([]*domain.Conversation, error) {
	return uc.conversationRepo.GetUserConversations(ctx, userID)
}

func (uc *ChatUseCase) SwitchAIModel(ctx context.Context, userID, conversationID int64, newModel string) error {
	// Get conversation
	conversation, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Verify user has access and it's an AI chat
	if !conversation.IsAIChat || (conversation.User1ID != userID && conversation.User2ID != userID) {
		return fmt.Errorf("invalid operation")
	}

	// Verify the new model is available
	_, err = uc.aiFactory.GetProvider(newModel)
	if err != nil {
		return fmt.Errorf("invalid AI model: %w", err)
	}

	// Update conversation
	conversation.ActiveAIModel = &newModel
	if err := uc.conversationRepo.Update(ctx, conversation); err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}
