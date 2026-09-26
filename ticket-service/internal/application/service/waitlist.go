package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type TicketsBack interface {
	TicketsBack(ctx context.Context, typeID int64, n int)
}

type WaitlistService struct {
	repo     outbound.WaitlistRepository
	events   outbound.EventRepository
	notifier *Notifier
	log      *slog.Logger
}

var (
	_ inbound.WaitlistUseCase = (*WaitlistService)(nil)
	_ TicketsBack             = (*WaitlistService)(nil)
)

func NewWaitlistService(repo outbound.WaitlistRepository, events outbound.EventRepository, n *Notifier, log *slog.Logger) *WaitlistService {
	return &WaitlistService{repo: repo, events: events, notifier: n, log: defaultLog(log)}
}

func (s *WaitlistService) Join(ctx context.Context, actor inbound.Principal, eventID, typeID int64) error {
	t, err := s.events.TicketType(ctx, typeID)
	if err != nil {
		return err
	}
	if t.EventID != eventID || !t.Active {
		return domain.ErrNotFound
	}
	e, err := s.events.Get(ctx, eventID)
	if err != nil {
		return err
	}
	if !e.OnSale(now()) {
		return domain.ErrNotFound
	}
	if t.Available > 0 {
		return fmt.Errorf("%w: tickets are still available, buy them now", domain.ErrConflict)
	}
	if err := s.repo.Join(ctx, typeID, actor.UserID); err != nil {
		if err == domain.ErrConflict {
			return fmt.Errorf("%w: you are already waiting for these tickets", domain.ErrConflict)
		}
		return err
	}
	return nil
}

func (s *WaitlistService) Leave(ctx context.Context, actor inbound.Principal, typeID int64) error {
	return s.repo.Leave(ctx, typeID, actor.UserID)
}

func (s *WaitlistService) Mine(ctx context.Context, actor inbound.Principal) ([]domain.WaitlistEntry, error) {
	return s.repo.Mine(ctx, actor.UserID)
}

func (s *WaitlistService) TicketsBack(ctx context.Context, typeID int64, n int) {
	if s == nil || n <= 0 {
		return
	}
	ctx = context.WithoutCancel(ctx)
	users, err := s.repo.Pop(ctx, typeID, min(n*3, 50))
	if err != nil {
		s.log.Warn("waiting list not read", "ticket_type", typeID, "err", err)
		return
	}
	if len(users) == 0 {
		return
	}
	t, err := s.events.TicketType(ctx, typeID)
	if err != nil {
		return
	}
	e, err := s.events.Get(ctx, t.EventID)
	if err != nil || !e.OnSale(now()) {
		return
	}
	for _, u := range users {
		s.notifier.Send(ctx, u, "tickets_available", "Tickets are back",
			fmt.Sprintf("%s tickets for %s are available again. Be quick!", t.Name, e.Title), 0, e.ID)
	}
}

func now() time.Time { return time.Now() }
