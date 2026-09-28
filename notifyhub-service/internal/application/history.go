package application

import (
	"context"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
)

type historyService struct {
	history port.HistoryRepository
}

func NewHistoryService(history port.HistoryRepository) port.HistoryUsecase {
	return &historyService{history: history}
}

func (s *historyService) List(ctx context.Context, f port.HistoryFilter) ([]*model.NotifyHistory, int64, error) {
	return s.history.List(ctx, f)
}
