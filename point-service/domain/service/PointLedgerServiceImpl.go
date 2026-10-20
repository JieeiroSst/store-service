package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/infrastructure/repository"
)

const (
	maxRewardPoints = 10_000_000
	maxPageSize     = 100
)

type PointLedgerServiceImpl struct {
	ledger repository.PointLedgerRepository
}

func NewPointLedgerServiceImpl(ledger repository.PointLedgerRepository) *PointLedgerServiceImpl {
	return &PointLedgerServiceImpl{ledger: ledger}
}

func (s *PointLedgerServiceImpl) Apply(ctx context.Context, event dto.RewardEventDTO) (bool, error) {
	event.EventID = strings.TrimSpace(event.EventID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.Source = strings.TrimSpace(event.Source)
	switch {
	case event.EventID == "" || len(event.EventID) > 64:
		return false, fmt.Errorf("%w: event_id must be 1-64 characters", ErrInvalidRewardEvent)
	case event.UserID == "" || len(event.UserID) > 64:
		return false, fmt.Errorf("%w: user_id must be 1-64 characters", ErrInvalidRewardEvent)
	case event.Points <= 0 || event.Points > maxRewardPoints:
		return false, fmt.Errorf("%w: points must be between 1 and %d", ErrInvalidRewardEvent, maxRewardPoints)
	}
	if event.Source == "" {
		event.Source = "unknown"
	}
	if len(event.Source) > 32 {
		event.Source = event.Source[:32]
	}
	if len(event.Title) > 255 {
		event.Title = event.Title[:255]
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}
	tx := event.ToEntity()
	return s.ledger.Credit(ctx, &tx)
}

func (s *PointLedgerServiceImpl) Balance(ctx context.Context, userID string) (*dto.PointAccountDTO, error) {
	a, err := s.ledger.GetAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := dto.AccountFromEntity(*a)
	return &out, nil
}

func (s *PointLedgerServiceImpl) Transactions(ctx context.Context, userID string, limit, offset int) (*dto.PointTransactionPageDTO, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	if offset < 0 {
		offset = 0
	}
	items, total, err := s.ledger.ListTransactions(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	page := &dto.PointTransactionPageDTO{Items: make([]dto.PointTransactionDTO, 0, len(items)), Total: total, Limit: limit, Offset: offset}
	for _, t := range items {
		page.Items = append(page.Items, dto.TransactionFromEntity(t))
	}
	return page, nil
}
