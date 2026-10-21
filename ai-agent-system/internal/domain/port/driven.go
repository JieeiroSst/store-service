package port

import (
	"context"

	"aiagentsystem/internal/domain/model"
)

type DeltaStream interface {
	Next() bool
	Current() string
	Result() model.Completion
	Err() error
	Close() error
}

type LLMClient interface {
	Complete(ctx context.Context, q model.Question) (model.Completion, error)
	Stream(ctx context.Context, q model.Question) (DeltaStream, error)
}

type HistoryRepository interface {
	SaveTurn(ctx context.Context, turn model.Turn) error
	ListByUser(ctx context.Context, userID string, limit int) ([]model.HistoryEntry, error)
}
