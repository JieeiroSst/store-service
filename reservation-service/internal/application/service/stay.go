package service

import (
	"context"
	"fmt"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

func (s *ReservationService) frontDesk(ctx context.Context, actor inbound.Principal, id int64) (domain.Reservation, error) {
	x, privileged, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	if !privileged {
		return domain.Reservation{}, domain.ErrForbidden
	}
	x = s.syncGateway(ctx, x)
	if x.Status != domain.ReservationPaid {
		return domain.Reservation{}, fmt.Errorf("%w: reservation is %s; only a paid one can be checked in or out", domain.ErrConflict, x.Status.Name())
	}
	return x, nil
}

func (s *ReservationService) setStay(ctx context.Context, actor inbound.Principal, x domain.Reservation, from, to domain.StayStatus) (domain.Reservation, error) {
	done, err := s.reservations.SetStay(ctx, x.ID, from, to, actor.UserID)
	if err != nil {
		if err == domain.ErrConflict {
			return domain.Reservation{}, fmt.Errorf("%w: the guest is %s; that step does not apply", domain.ErrConflict, x.Stay.Name())
		}
		return domain.Reservation{}, err
	}
	return done, nil
}

func (s *ReservationService) CheckIn(ctx context.Context, actor inbound.Principal, id int64) (domain.Reservation, error) {
	x, err := s.frontDesk(ctx, actor, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	today := s.today()
	if today.Before(x.Start) {
		return domain.Reservation{}, fmt.Errorf("%w: the stay starts on %s", domain.ErrConflict, x.Start.Format(dateOnly))
	}
	if !today.Before(x.End) {
		return domain.Reservation{}, fmt.Errorf("%w: the stay ended on %s", domain.ErrConflict, x.End.Format(dateOnly))
	}
	return s.setStay(ctx, actor, x, domain.StayNotArrived, domain.StayCheckedIn)
}

func (s *ReservationService) CheckOut(ctx context.Context, actor inbound.Principal, id int64) (domain.Reservation, error) {
	x, err := s.frontDesk(ctx, actor, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	done, err := s.setStay(ctx, actor, x, domain.StayCheckedIn, domain.StayCheckedOut)
	if err == nil {
		s.opts.Notifier.Send(ctx, x.GuestID, "reservation.checked_out", "Thank you for staying with us",
			fmt.Sprintf("We hope you enjoyed your stay %s. Please tell others about it by rating the hotel.", stayText(x)), x.ID, x.HotelID)
	}
	return done, err
}

func (s *ReservationService) NoShow(ctx context.Context, actor inbound.Principal, id int64) (domain.Reservation, error) {
	x, err := s.frontDesk(ctx, actor, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	if !s.today().After(x.Start) {
		return domain.Reservation{}, fmt.Errorf("%w: the guest can still arrive on %s", domain.ErrConflict, x.Start.Format(dateOnly))
	}
	return s.setStay(ctx, actor, x, domain.StayNotArrived, domain.StayNoShow)
}

func (s *ReservationService) History(ctx context.Context, actor inbound.Principal, id int64) ([]inbound.HistoryEntry, error) {
	x, _, err := s.load(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	events, err := s.reservations.Events(ctx, x.ID)
	if err != nil {
		return nil, err
	}
	out := make([]inbound.HistoryEntry, len(events))
	for i, e := range events {
		by := "hotel"
		switch e.Actor {
		case 0:
			by = "system"
		case x.GuestID:
			by = "guest"
		}
		out[i] = inbound.HistoryEntry{At: e.At, By: by, Event: e.Event, From: e.FromStatus, To: e.ToStatus, Note: e.Note}
	}
	return out, nil
}
