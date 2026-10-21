package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"aiagentsystem/internal/domain/model"
	"aiagentsystem/internal/domain/port"
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

type HistoryRepository struct {
	db *sql.DB
}

var _ port.HistoryRepository = (*HistoryRepository)(nil)

func NewHistoryRepository(db *sql.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

func (r *HistoryRepository) Migrate(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("history: failed to ensure schema: %w", err)
	}
	return nil
}

func (r *HistoryRepository) SaveTurn(ctx context.Context, turn model.Turn) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("history: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	const insert = `INSERT INTO chat_messages (user_id, role, content, source) VALUES (?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, insert, turn.UserID, model.RoleUser, turn.Question, nil); err != nil {
		return fmt.Errorf("history: failed to save question: %w", err)
	}
	if _, err := tx.ExecContext(ctx, insert, turn.UserID, model.RoleAssistant, turn.Answer, string(turn.Source)); err != nil {
		return fmt.Errorf("history: failed to save answer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("history: failed to commit: %w", err)
	}
	return nil
}

func (r *HistoryRepository) ListByUser(ctx context.Context, userID string, limit int) ([]model.HistoryEntry, error) {
	const query = `
		SELECT id, user_id, role, content, COALESCE(source, ''), created_at
		FROM chat_messages
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("history: failed to list messages: %w", err)
	}
	defer rows.Close()

	var entries []model.HistoryEntry
	for rows.Next() {
		var e model.HistoryEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.Role, &e.Content, &e.Source, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("history: failed to scan message: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("history: failed to read messages: %w", err)
	}
	return entries, nil
}
