package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

const maxStayNights = 90

type BookingOptions struct {
	HoldTTL time.Duration
	Log     *slog.Logger
}

type BookingService struct {
	bookings  outbound.BookingRepository
	homestays outbound.HomestayRepository
	rail      *PaymentRail
	opts      BookingOptions
	now       func() time.Time
}

var _ inbound.BookingUseCase = (*BookingService)(nil)

func NewBookingService(b outbound.BookingRepository, h outbound.HomestayRepository, rail *PaymentRail, opts BookingOptions) *BookingService {
	if opts.Log == nil {
		opts.Log = slog.Default()
	}
	if opts.HoldTTL <= 0 {
		opts.HoldTTL = 15 * time.Minute
	}
	return &BookingService{bookings: b, homestays: h, rail: rail, opts: opts, now: time.Now}
}

func (s *BookingService) Book(ctx context.Context, actor inbound.Principal, c inbound.BookCommand) (domain.Booking, error) {
	today := s.now().UTC().Truncate(24 * time.Hour)
	if c.CheckIn.Before(today) {
		return domain.Booking{}, fmt.Errorf("%w: check-in is in the past", domain.ErrInvalid)
	}
	if !c.CheckOut.After(c.CheckIn) {
		return domain.Booking{}, fmt.Errorf("%w: check-out must be after check-in", domain.ErrInvalid)
	}
	if c.CheckOut.Sub(c.CheckIn) > maxStayNights*24*time.Hour {
		return domain.Booking{}, fmt.Errorf("%w: stay cannot exceed %d nights", domain.ErrInvalid, maxStayNights)
	}
	if c.Guests <= 0 {
		return domain.Booking{}, fmt.Errorf("%w: guests must be positive", domain.ErrInvalid)
	}

	h, err := s.homestays.Get(ctx, c.HomestayID)
	if err != nil {
		return domain.Booking{}, err
	}
	if h.Status != domain.HomestayActive {
		return domain.Booking{}, domain.ErrNotFound
	}
	if h.HostID != 0 && h.HostID == actor.UserID {
		return domain.Booking{}, fmt.Errorf("%w: you cannot book your own homestay", domain.ErrForbidden)
	}
	if c.Guests > h.Guests {
		return domain.Booking{}, fmt.Errorf("%w: homestay sleeps at most %d guests", domain.ErrInvalid, h.Guests)
	}

	reqID := strings.TrimSpace(c.RequestID)
	if reqID == "" {
		if reqID, err = randomID(); err != nil {
			return domain.Booking{}, err
		}
	}
	currency := strings.ToUpper(strings.TrimSpace(c.Currency))
	if currency == "" {
		currency = "VND"
	}
	return s.bookings.Reserve(ctx, outbound.ReserveParams{
		UserID: actor.UserID, HomestayID: c.HomestayID,
		CheckIn: c.CheckIn, CheckOut: c.CheckOut, Guests: c.Guests,
		Currency: currency, Note: c.Note, RequestID: reqID,
		ExpiresAt: s.now().Add(s.opts.HoldTTL),
	})
}

func (s *BookingService) load(ctx context.Context, actor inbound.Principal, id int64) (b domain.Booking, privileged bool, err error) {
	b, err = s.bookings.Get(ctx, id)
	if err != nil {
		return b, false, err
	}
	allowed, privileged := stayAccess(ctx, s.homestays, actor, b.UserID, b.HomestayID)
	if !allowed {
		return domain.Booking{}, false, domain.ErrNotFound
	}
	return b, privileged, nil
}

func (s *BookingService) Get(ctx context.Context, actor inbound.Principal, id int64) (domain.Booking, error) {
	b, _, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Booking{}, err
	}
	return s.syncGateway(ctx, b), nil
}

func (s *BookingService) Cancel(ctx context.Context, actor inbound.Principal, id int64) (domain.Booking, error) {
	b, privileged, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Booking{}, err
	}
	b = s.syncGateway(ctx, b)
	if b.Status == domain.BookingCancelled {
		return b, nil
	}
	if !b.CheckIn.After(s.now().UTC().Truncate(24*time.Hour)) && !privileged {
		return domain.Booking{}, fmt.Errorf("%w: stay has already started", domain.ErrConflict)
	}

	if err := s.rail.refund(ctx, b.PaymentMethod, b.PaymentRef, fmt.Sprintf("booking #%d cancelled", b.ID), b.Status != domain.BookingConfirmed); err != nil {
		return domain.Booking{}, err
	}
	return s.bookings.Cancel(ctx, b.ID, actor.UserID)
}

func (s *BookingService) List(ctx context.Context, actor inbound.Principal, userID int64, asHost bool, limit, offset int) ([]domain.Booking, error) {
	var hostID int64
	switch {
	case asHost:
		userID, hostID = 0, actor.UserID
	case !actor.IsAdmin():
		userID = actor.UserID
	}
	limit, offset = page(limit, offset)
	return s.bookings.List(ctx, userID, hostID, limit, offset)
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
