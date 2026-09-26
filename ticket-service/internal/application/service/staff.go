package service

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type StaffService struct {
	events outbound.EventRepository
	staff  outbound.StaffRepository
}

var _ inbound.StaffUseCase = (*StaffService)(nil)

func NewStaffService(events outbound.EventRepository, staff outbound.StaffRepository) *StaffService {
	return &StaffService{events: events, staff: staff}
}

func (s *StaffService) manage(ctx context.Context, actor inbound.Principal, eventID int64) (domain.Event, error) {
	e, err := s.events.Get(ctx, eventID)
	if err != nil {
		return e, err
	}
	if !canManage(actor, e) {
		return e, notYours(e)
	}
	return e, nil
}

func (s *StaffService) Add(ctx context.Context, actor inbound.Principal, eventID, userID int64) error {
	e, err := s.manage(ctx, actor, eventID)
	if err != nil {
		return err
	}
	switch {
	case userID <= 0:
		return invalid("user_id is required")
	case userID == e.OrganizerID:
		return invalid("the organizer can already scan tickets")
	case e.Status == domain.EventCancelled:
		return fmt.Errorf("%w: the event was cancelled", domain.ErrConflict)
	}
	return s.staff.Add(ctx, eventID, userID, actor.UserID)
}

func (s *StaffService) Remove(ctx context.Context, actor inbound.Principal, eventID, userID int64) error {
	if _, err := s.manage(ctx, actor, eventID); err != nil {
		return err
	}
	return s.staff.Remove(ctx, eventID, userID)
}

func (s *StaffService) List(ctx context.Context, actor inbound.Principal, eventID int64) ([]domain.StaffMember, error) {
	if _, err := s.manage(ctx, actor, eventID); err != nil {
		return nil, err
	}
	return s.staff.List(ctx, eventID)
}

func (s *StaffService) Events(ctx context.Context, actor inbound.Principal) ([]domain.Event, error) {
	return s.staff.EventsFor(ctx, actor.UserID)
}
