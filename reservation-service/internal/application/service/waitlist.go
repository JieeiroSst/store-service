package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

const (
	maxWaitingPerGuest = 5
	promoteBatch       = 50
)

type WaitlistOptions struct {
	OfferTTL   time.Duration
	MaxPending int
	Log        *slog.Logger
	Notifier   *Notifier
	SoldOut    *SoldOutCache
}

type WaitlistService struct {
	list         outbound.WaitlistRepository
	reservations outbound.ReservationRepository
	hotels       outbound.HotelRepository
	opts         WaitlistOptions
	now          func() time.Time
}

var _ inbound.WaitlistUseCase = (*WaitlistService)(nil)

func NewWaitlistService(l outbound.WaitlistRepository, r outbound.ReservationRepository, h outbound.HotelRepository, opts WaitlistOptions) *WaitlistService {
	opts.Log = defaultLog(opts.Log)
	if opts.OfferTTL <= 0 {
		opts.OfferTTL = 15 * time.Minute
	}
	return &WaitlistService{list: l, reservations: r, hotels: h, opts: opts, now: time.Now}
}

func (s *WaitlistService) today() time.Time { return s.now().UTC().Truncate(24 * time.Hour) }

func (s *WaitlistService) Join(ctx context.Context, actor inbound.Principal, c inbound.WaitCommand) (domain.WaitlistEntry, error) {
	invalid := func(format string, a ...any) (domain.WaitlistEntry, error) {
		return domain.WaitlistEntry{}, fmt.Errorf("%w: "+format, append([]any{domain.ErrInvalid}, a...)...)
	}
	if c.Start.Before(s.today()) {
		return invalid("check-in is in the past")
	}
	if !c.End.After(c.Start) {
		return invalid("check-out must be after check-in")
	}
	if c.End.Sub(c.Start) > maxStayNights*24*time.Hour {
		return invalid("a stay cannot exceed %d nights", maxStayNights)
	}
	if c.Rooms < 1 || c.Rooms > maxRooms {
		return invalid("rooms must be between 1 and %d", maxRooms)
	}
	if c.Adults < 1 || c.Children < 0 {
		return invalid("at least one adult is required")
	}
	h, err := s.hotels.Get(ctx, c.HotelID)
	if err != nil {
		return domain.WaitlistEntry{}, err
	}
	if h.Status != domain.HotelActive {
		return domain.WaitlistEntry{}, domain.ErrNotFound
	}
	if h.OwnerID != 0 && h.OwnerID == actor.UserID {
		return domain.WaitlistEntry{}, fmt.Errorf("%w: you cannot reserve a room at your own hotel", domain.ErrForbidden)
	}
	var rt *domain.RoomType
	for i := range h.RoomTypes {
		if h.RoomTypes[i].ID == c.RoomTypeID && h.RoomTypes[i].Active {
			rt = &h.RoomTypes[i]
		}
	}
	if rt == nil {
		return domain.WaitlistEntry{}, fmt.Errorf("%w: room type is not offered by this hotel", domain.ErrNotFound)
	}
	if c.Adults+c.Children > c.Rooms*rt.Capacity {
		return invalid("%d room(s) of %s sleep at most %d guests", c.Rooms, rt.Name, c.Rooms*rt.Capacity)
	}
	left, err := s.reservations.RoomsLeft(ctx, h.ID, rt.ID, c.Start, c.End)
	if err != nil {
		return domain.WaitlistEntry{}, err
	}
	if left >= c.Rooms {
		return domain.WaitlistEntry{}, fmt.Errorf("%w: rooms are available for those nights, reserve them instead of waiting", domain.ErrConflict)
	}
	return s.list.Join(ctx, domain.WaitlistEntry{HotelID: h.ID, RoomTypeID: rt.ID, GuestID: actor.UserID, Start: c.Start, End: c.End,
		Rooms: c.Rooms, Adults: c.Adults, Children: c.Children}, maxWaitingPerGuest)
}

func (s *WaitlistService) Leave(ctx context.Context, actor inbound.Principal, id int64) error {
	return s.list.Leave(ctx, actor.UserID, id)
}

func (s *WaitlistService) List(ctx context.Context, actor inbound.Principal, afterID int64, limit int) (domain.Page[domain.WaitlistEntry], error) {
	limit, _ = page(limit, 0)
	rows, err := s.list.List(ctx, actor.UserID, afterID, limit+1)
	if err != nil {
		return domain.Page[domain.WaitlistEntry]{}, err
	}
	return domain.PageOf(rows, limit, func(e domain.WaitlistEntry) int64 { return e.ID }), nil
}

func (s *WaitlistService) Promote(ctx context.Context, roomTypeID int64) (int, error) {
	waiting, err := s.list.Waiting(ctx, roomTypeID, s.today(), promoteBatch)
	if err != nil || len(waiting) == 0 {
		return 0, err
	}
	offered := 0
	for _, e := range waiting {
		left, err := s.reservations.RoomsLeft(ctx, e.HotelID, e.RoomTypeID, e.Start, e.End)
		if err != nil {
			return offered, err
		}
		if left < e.Rooms {
			continue
		}
		h, err := s.hotels.Get(ctx, e.HotelID)
		if err != nil || h.Status != domain.HotelActive {
			continue
		}

		x, err := s.reservations.Reserve(ctx, outbound.ReserveParams{
			GuestID: e.GuestID, HotelID: e.HotelID, RoomTypeID: e.RoomTypeID, Start: e.Start, End: e.End,
			Rooms: e.Rooms, Adults: e.Adults, Children: e.Children, Currency: h.Currency,
			RequestID: fmt.Sprintf("waitlist-%d", e.ID), ExpiresAt: s.now().Add(s.opts.OfferTTL),
			AddonsTotal: "0.0000", MaxPending: s.opts.MaxPending, Actor: 0,
		})
		switch {
		case errors.Is(err, domain.ErrDatesUnavailable):
			continue 
		case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrBusy):
			continue 
		case err != nil:
			s.opts.Log.Error("waiting list offer failed", "entry", e.ID, "err", err)
			continue
		}
		result, err := s.list.MarkOffered(ctx, e.ID, x.ID)
		if err != nil {
			s.opts.Log.Error("waiting list offer not recorded", "entry", e.ID, "reservation", x.ID, "err", err)
			continue
		}
		switch result {
		case outbound.OfferAlready:
			continue
		case outbound.OfferGone:
			if _, err := s.reservations.Transition(ctx, outbound.TransitionParams{ID: x.ID, From: []domain.ReservationStatus{domain.ReservationPending},
				To: domain.ReservationCanceled, Note: "the guest left the waiting list"}); err == nil {
				s.opts.SoldOut.Release(e.RoomTypeID)
			}
			continue
		}
		offered++
		s.opts.SoldOut.Release(e.RoomTypeID) 
		s.opts.Notifier.Send(ctx, e.GuestID, "waitlist.offer",
			fmt.Sprintf("A room is available at %s", h.Name),
			fmt.Sprintf("Rooms you were waiting for (%s to %s) are yours for %d minutes. Pay reservation #%d to keep them.",
				e.Start.Format(dateOnly), e.End.Format(dateOnly), int(s.opts.OfferTTL.Minutes()), x.ID), x.ID, h.ID)
		s.opts.Notifier.Send(ctx, h.OwnerID, "reservation.created",
			fmt.Sprintf("New reservation #%d at %s", x.ID, h.Name),
			fmt.Sprintf("A guest from the waiting list is holding rooms for %s to %s. It is pending until they pay.",
				e.Start.Format(dateOnly), e.End.Format(dateOnly)), x.ID, h.ID)
	}
	return offered, nil
}

func (s *WaitlistService) PromoteAll(ctx context.Context) (int, error) {
	if _, err := s.list.ExpireStale(ctx, s.today()); err != nil {
		return 0, err
	}
	types, err := s.list.RoomTypesWaiting(ctx, s.today(), 200)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, id := range types {
		n, err := s.Promote(ctx, id)
		if err != nil {
			s.opts.Log.Error("waiting list round failed", "room_type", id, "err", err)
			continue
		}
		total += n
	}
	return total, nil
}
