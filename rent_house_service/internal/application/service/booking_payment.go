package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

const expireBatch = 100

func (s *BookingService) Pay(ctx context.Context, actor inbound.Principal, id int64, cmd inbound.PayCommand) (domain.Booking, error) {
	b, err := s.bookings.Get(ctx, id)
	if err != nil {
		return domain.Booking{}, err
	}
	if b.UserID != actor.UserID {
		return domain.Booking{}, domain.ErrNotFound
	}
	switch b.Status {
	case domain.BookingConfirmed:
		return b, nil // already paid
	case domain.BookingCancelled:
		return domain.Booking{}, fmt.Errorf("%w: booking is cancelled or its hold expired", domain.ErrConflict)
	}
	if b.ExpiresAt != nil && s.now().After(*b.ExpiresAt) {
		return domain.Booking{}, fmt.Errorf("%w: booking hold expired", domain.ErrConflict)
	}

	amount, err := domain.MinorUnits(b.TotalAmount, b.Currency)
	if err != nil {
		return domain.Booking{}, err
	}
	if amount == 0 {
		return s.confirm(ctx, b, amount, domain.MethodWallet, "free")
	}

	switch cmd.Method {
	case domain.MethodWallet:
		return s.payWithWallet(ctx, actor, b, amount)
	case domain.MethodGateway:
		return s.payWithGateway(ctx, actor, b, amount, cmd.Provider)
	default:
		return domain.Booking{}, fmt.Errorf("%w: method must be %q or %q", domain.ErrInvalid, domain.MethodWallet, domain.MethodGateway)
	}
}

func (s *BookingService) confirm(ctx context.Context, b domain.Booking, amount int64, m domain.PaymentMethod, ref string) (domain.Booking, error) {
	paid, err := s.bookings.MarkPaid(ctx, b.ID, m, ref)
	if err == nil {
		s.rail.earn(ctx, b.UserID, fmt.Sprintf("rent-house-booking-%d", b.ID), amount)
	}
	return paid, err
}

func (s *BookingService) payWithWallet(ctx context.Context, actor inbound.Principal, b domain.Booking, amount int64) (domain.Booking, error) {
	h, err := s.homestays.Get(ctx, b.HomestayID)
	if err != nil {
		return domain.Booking{}, err
	}
	transferID, err := s.rail.chargeWallet(ctx, actor.UserID, h.WalletID, b.Currency, amount,
		fmt.Sprintf("rent-house-booking-%d", b.ID), fmt.Sprintf("Booking #%d", b.ID))
	if err != nil {
		return domain.Booking{}, err
	}
	paid, err := s.confirm(ctx, b, amount, domain.MethodWallet, transferID)
	if errors.Is(err, domain.ErrConflict) {
		s.rail.undoWallet(ctx, transferID, fmt.Sprintf("booking #%d no longer payable", b.ID))
		return domain.Booking{}, fmt.Errorf("%w: booking is no longer payable, payment refunded", domain.ErrConflict)
	}
	return paid, err
}

func (s *BookingService) payWithGateway(ctx context.Context, actor inbound.Principal, b domain.Booking, amount int64, provider string) (domain.Booking, error) {
	existing := ""
	if b.PaymentMethod == domain.MethodGateway {
		existing = b.PaymentRef
	}

	p, fresh, err := s.rail.gatewayPay(ctx, existing, outbound.CreateGatewayPayment{
		Provider: strings.ToLower(strings.TrimSpace(provider)), Amount: amount, Currency: strings.ToUpper(b.Currency),
		PayerEmail: actor.Email, Description: fmt.Sprintf("Booking #%d", b.ID),
		IdempotencyKey: fmt.Sprintf("rent-house-booking-%d-%d", b.ID, b.Version),
	})
	if err != nil {
		return domain.Booking{}, err
	}
	if !fresh {
		if p.Status == domain.GatewayCaptured {
			return s.confirm(ctx, b, amount, domain.MethodGateway, existing)
		}
		return b, nil
	}
	ref := strconv.FormatInt(p.ID, 10)
	if b, err = s.bookings.SetPaymentAttempt(ctx, b.ID, domain.MethodGateway, ref); err != nil {
		return domain.Booking{}, err
	}
	switch p.Status {
	case domain.GatewayCaptured:
		return s.confirm(ctx, b, amount, domain.MethodGateway, ref)
	case domain.GatewayFailed:
		return domain.Booking{}, domain.ErrPaymentFailed
	}
	return b, nil
}

func (s *BookingService) syncGateway(ctx context.Context, b domain.Booking) domain.Booking {
	if b.Status != domain.BookingPending || b.PaymentMethod != domain.MethodGateway || b.PaymentRef == "" {
		return b
	}
	p, ok := s.rail.gatewayStatus(ctx, b.PaymentRef)
	if !ok || p.Status != domain.GatewayCaptured {
		return b
	}
	amount, _ := domain.MinorUnits(b.TotalAmount, b.Currency)
	if paid, err := s.confirm(ctx, b, amount, domain.MethodGateway, b.PaymentRef); err == nil {
		return paid
	}
	return b
}

func (s *BookingService) ReleaseExpired(ctx context.Context) (int, error) {
	now := s.now()
	expired, err := s.bookings.ListExpired(ctx, now, expireBatch)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, b := range expired {
		if b.PaymentMethod == domain.MethodGateway && b.PaymentRef != "" {
			switch s.rail.judgeStale(ctx, b.PaymentRef, b.ExpiresAt, s.opts.HoldTTL, now) {
			case staleKeep:
				continue
			case stalePaid:
				amount, _ := domain.MinorUnits(b.TotalAmount, b.Currency)
				if _, err := s.confirm(ctx, b, amount, domain.MethodGateway, b.PaymentRef); err != nil {
					s.opts.Log.Error("confirm captured payment", "booking", b.ID, "err", err)
				}
				continue
			}
		}
		ok, err := s.bookings.Expire(ctx, b.ID)
		if err != nil {
			s.opts.Log.Error("expire booking", "booking", b.ID, "err", err)
			continue
		}
		if ok {
			released++
		}
	}
	return released, nil
}
