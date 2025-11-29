package database

import (
	"context"
	"database/sql"
	"fmt"

	"chatbot-system/internal/domain"
)

type MySQLConversationRepository struct {
	db *sql.DB
}

func NewMySQLConversationRepository(db *sql.DB) *MySQLConversationRepository {
	return &MySQLConversationRepository{db: db}
}

func (r *MySQLConversationRepository) Create(ctx context.Context, conv *domain.Conversation) error {
	query := `
		INSERT INTO conversations (user1_id, user2_id, is_ai_chat, active_ai_model, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`
	result, err := r.db.ExecContext(ctx, query, conv.User1ID, conv.User2ID, conv.IsAIChat, conv.ActiveAIModel)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	conv.ID = id
	return nil
}

func (r *MySQLConversationRepository) GetByID(ctx context.Context, id int64) (*domain.Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, is_ai_chat, active_ai_model, created_at, updated_at
		FROM conversations
		WHERE id = ?
	`
	conv := &domain.Conversation{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID,
		&conv.User1ID,
		&conv.User2ID,
		&conv.IsAIChat,
		&conv.ActiveAIModel,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conv, nil
}

func (r *MySQLConversationRepository) GetByUsers(ctx context.Context, user1ID, user2ID int64) (*domain.Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, is_ai_chat, active_ai_model, created_at, updated_at
		FROM conversations
		WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)
		LIMIT 1
	`
	conv := &domain.Conversation{}
	err := r.db.QueryRowContext(ctx, query, user1ID, user2ID, user2ID, user1ID).Scan(
		&conv.ID,
		&conv.User1ID,
		&conv.User2ID,
		&conv.IsAIChat,
		&conv.ActiveAIModel,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No conversation found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conv, nil
}

func (r *MySQLConversationRepository) GetUserConversations(ctx context.Context, userID int64) ([]*domain.Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, is_ai_chat, active_ai_model, created_at, updated_at
		FROM conversations
		WHERE user1_id = ? OR user2_id = ?
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*domain.Conversation
	for rows.Next() {
		conv := &domain.Conversation{}
		if err := rows.Scan(
			&conv.ID,
			&conv.User1ID,
			&conv.User2ID,
			&conv.IsAIChat,
			&conv.ActiveAIModel,
			&conv.CreatedAt,
			&conv.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (r *MySQLConversationRepository) Update(ctx context.Context, conv *domain.Conversation) error {
	query := `
		UPDATE conversations
		SET active_ai_model = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, conv.ActiveAIModel, conv.ID)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// Message Repository

type MySQLMessageRepository struct {
	db *sql.DB
}

func NewMySQLMessageRepository(db *sql.DB) *MySQLMessageRepository {
	return &MySQLMessageRepository{db: db}
}

func (r *MySQLMessageRepository) Create(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (conversation_id, sender_id, content, message_type, ai_model, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`
	result, err := r.db.ExecContext(ctx, query, message.ConversationID, message.SenderID, message.Content, message.MessageType, message.AIModel)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	message.ID = id
	return nil
}

func (r *MySQLMessageRepository) GetByID(ctx context.Context, id int64) (*domain.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, content, message_type, ai_model, created_at
		FROM messages
		WHERE id = ?
	`
	msg := &domain.Message{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msg.Content,
		&msg.MessageType,
		&msg.AIModel,
		&msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return msg, nil
}

func (r *MySQLMessageRepository) GetConversationMessages(ctx context.Context, conversationID int64, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, content, message_type, ai_model, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		if err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.Content,
			&msg.MessageType,
			&msg.AIModel,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *MySQLMessageRepository) GetConversationHistory(ctx context.Context, conversationID int64) ([]*domain.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, content, message_type, ai_model, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		if err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.Content,
			&msg.MessageType,
			&msg.AIModel,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
