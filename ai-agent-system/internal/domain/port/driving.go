package port

import (
	"context"

	"aiagentsystem/internal/domain/model"
)

type ChatService interface {
	Answer(ctx context.Context, q model.Question) (model.Answer, error)
	AnswerStream(ctx context.Context, q model.Question, onDelta func(delta string)) (model.Answer, error)
}

type HistoryService interface {
	List(ctx context.Context, userID string, limit int) ([]model.HistoryEntry, error)
}
