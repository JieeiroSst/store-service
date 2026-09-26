package service

import (
	"context"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type WishlistService struct {
	repo   outbound.WishlistRepository
	events outbound.EventRepository
}

var _ inbound.WishlistUseCase = (*WishlistService)(nil)

func NewWishlistService(repo outbound.WishlistRepository, events outbound.EventRepository) *WishlistService {
	return &WishlistService{repo: repo, events: events}
}

func (s *WishlistService) Add(ctx context.Context, a inbound.Principal, eventID int64) error {
	e, err := s.events.Get(ctx, eventID)
	if err != nil {
		return err
	}
	if e.Status != domain.EventPublished {
		return domain.ErrNotFound
	}
	return s.repo.Add(ctx, a.UserID, eventID)
}

func (s *WishlistService) Remove(ctx context.Context, a inbound.Principal, eventID int64) error {
	return s.repo.Remove(ctx, a.UserID, eventID)
}

func (s *WishlistService) List(ctx context.Context, a inbound.Principal, limit, offset int) ([]domain.Event, error) {
	return s.repo.List(ctx, a.UserID, page(limit), max(offset, 0))
}
