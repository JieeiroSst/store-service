package history

import (
	"context"
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS chat_messages (
	id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
	user_id VARCHAR(255) NOT NULL,
	role VARCHAR(16) NOT NULL,
	content TEXT NOT NULL,
	source VARCHAR(20) NULL,
	created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
	INDEX idx_user_created (user_id, created_at)
)`

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("history: failed to ensure schema: %w", err)
	}
	return nil
}

func (s *Store) SaveTurn(ctx context.Context, userID, question, answer, source string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("history: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	const insert = `INSERT INTO chat_messages (user_id, role, content, source) VALUES (?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, insert, userID, "user", question, nil); err != nil {
		return fmt.Errorf("history: failed to save question: %w", err)
	}
	if _, err := tx.ExecContext(ctx, insert, userID, "assistant", answer, source); err != nil {
		return fmt.Errorf("history: failed to save answer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("history: failed to commit: %w", err)
	}
	return nil
}

func (s *Store) ListByUser(ctx context.Context, userID string, limit int) ([]Message, error) {
	const query = `
		SELECT id, user_id, role, content, COALESCE(source, ''), created_at
		FROM chat_messages
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?`

	rows, err := s.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("history: failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.Source, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("history: failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("history: failed to read messages: %w", err)
	}
	return messages, nil
}
