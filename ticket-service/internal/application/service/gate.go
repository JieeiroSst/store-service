package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const maxBatchScans = 500

type GateService struct {
	events  outbound.EventRepository
	tickets outbound.TicketRepository
	staff   outbound.StaffRepository
	now     func() time.Time
}

var _ inbound.GateUseCase = (*GateService)(nil)

func NewGateService(events outbound.EventRepository, tickets outbound.TicketRepository, staff outbound.StaffRepository) *GateService {
	return &GateService{events: events, tickets: tickets, staff: staff, now: time.Now}
}

func (s *GateService) authorize(ctx context.Context, actor inbound.Principal, eventID int64) (domain.Event, error) {
	e, err := s.events.Get(ctx, eventID)
	if err != nil {
		return e, err
	}
	if !canScan(ctx, s.staff, actor, e) {
		return e, notYours(e)
	}
	if e.Status != domain.EventPublished && e.Status != domain.EventCancelled {
		return e, domain.ErrNotFound
	}
	return e, nil
}

func (s *GateService) CheckIn(ctx context.Context, actor inbound.Principal, eventID int64, code string) (domain.Ticket, error) {
	e, err := s.authorize(ctx, actor, eventID)
	if err != nil {
		return domain.Ticket{}, err
	}
	if code = normalizeCode(code); code == "" || len(code) > 64 {
		return domain.Ticket{}, invalid("code is required")
	}
	now := s.now()
	if e.Status == domain.EventCancelled {
		return domain.Ticket{}, fmt.Errorf("%w: the event was cancelled", domain.ErrConflict)
	}
	t, err := s.tickets.ByCode(ctx, eventID, code)
	if err != nil {
		return domain.Ticket{}, err
	}
	sess := showtime(e, t.SessionID)
	switch {
	case sess.Status == domain.SessionCancelled:
		return t, fmt.Errorf("%w: this showtime was cancelled", domain.ErrConflict)
	case now.After(sess.EndsAt):
		return t, fmt.Errorf("%w: this showtime is over", domain.ErrConflict)
	}
	return s.tickets.CheckIn(ctx, eventID, code, actor.UserID, now)
}

func (s *GateService) BatchCheckIn(ctx context.Context, actor inbound.Principal, eventID int64, items []inbound.ScanItem) ([]inbound.ScanResult, error) {
	e, err := s.authorize(ctx, actor, eventID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || len(items) > maxBatchScans {
		return nil, invalid("send between 1 and %d scans", maxBatchScans)
	}
	now := s.now()
	out := make([]inbound.ScanResult, len(items))
	for i, it := range items {
		out[i].Code = it.Code
		code := normalizeCode(it.Code)
		if code == "" || len(code) > 64 {
			out[i].Result = "not_found"
			continue
		}
		known, err := s.tickets.ByCode(ctx, eventID, code)
		if errors.Is(err, domain.ErrNotFound) {
			out[i].Result = "not_found"
			continue
		}
		if err != nil {
			return out[:i], err
		}
		sess := showtime(e, known.SessionID)
		at := now
		if it.ScannedAt != nil && !it.ScannedAt.After(now) && it.ScannedAt.After(sess.StartsAt.Add(-24*time.Hour)) {
			at = *it.ScannedAt
		}
		if e.Status == domain.EventCancelled || sess.Status == domain.SessionCancelled || at.After(sess.EndsAt) || now.After(sess.EndsAt.Add(expiryGrace)) {
			out[i].Result = "event_over"
			continue
		}
		t, err := s.tickets.CheckIn(ctx, eventID, code, actor.UserID, at)
		out[i].Ticket = t
		switch {
		case err == nil:
			out[i].Result = "admitted"
		case errors.Is(err, domain.ErrAlreadyCheckedIn):
			out[i].Result = "already_used"
		case errors.Is(err, domain.ErrNotFound):
			out[i].Result = "not_found"
		case errors.Is(err, domain.ErrConflict) && t.ID != 0:
			out[i].Result = t.Status.Name()
		default:
			return out[:i], err
		}
	}
	return out, nil
}
