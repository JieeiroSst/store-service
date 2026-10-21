package application

import (
	"context"

	"aiagentsystem/internal/domain/model"
	"aiagentsystem/internal/domain/port"
)

const (
	DefaultHistoryLimit = 50
	MaxHistoryLimit     = 200
)

type HistoryService struct {
	repo port.HistoryRepository
}

var _ port.HistoryService = (*HistoryService)(nil)

func NewHistoryService(repo port.HistoryRepository) *HistoryService {
	return &HistoryService{repo: repo}
}

func (s *HistoryService) List(ctx context.Context, userID string, limit int) ([]model.HistoryEntry, error) {
	if userID == "" {
		return nil, model.BadRequest("user id is required")
	}
	if limit < 0 {
		return nil, model.BadRequest("limit must be a positive integer")
	}
	if limit == 0 {
		limit = DefaultHistoryLimit
	}
	if limit > MaxHistoryLimit {
		limit = MaxHistoryLimit
	}

	entries, err := s.repo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, model.AsError(err)
	}
	return entries, nil
}
