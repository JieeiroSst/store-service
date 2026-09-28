package application

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/google/uuid"
)

const (
	reconcileBatch   = 50
	metadataWalletID = "wallet_id"
	metadataUnclaim  = "unclaimed"
)

var errAlreadySettled = errors.New("payment already settled")

func (s *vendingService) settle(ctx context.Context, paymentID, txID, machineID string) (*domain.Checkout, error) {
	var (
		payment   *domain.Payment
		order     *domain.Order
		unclaimed bool
	)
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		p, err := s.payments.GetForUpdate(ctx, paymentID)
		if err != nil {
			return err
		}
		r, err := s.reservations.GetForUpdate(ctx, p.ReservationID)
		if err != nil {
			return err
		}
		lateCharge := p.Status == domain.PaymentFailed && txID != ""
		if p.Status != domain.PaymentPending && !lateCharge {
			return errAlreadySettled
		}
		now := s.now().UTC()
		p.TransactionID = txID
		p.Status = domain.PaymentCompleted
		p.CompletedAt = &now
		p.UpdatedAt = now
		payment = p

		if lateCharge || r.Status != domain.ReservationPending {
			unclaimed = true
			p.Metadata[metadataUnclaim] = "true"
			return s.payments.Update(ctx, p)
		}

		r.Status = domain.ReservationConfirmed
		r.UpdatedAt = now
		if err := s.reservations.Update(ctx, r); err != nil {
			return err
		}
		if err := s.payments.Update(ctx, p); err != nil {
			return err
		}
		order = &domain.Order{
			ID:            uuid.NewString(),
			SessionID:     r.SessionID,
			ReservationID: r.ID,
			PaymentID:     p.ID,
			Status:        domain.OrderProcessing,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := s.orders.Create(ctx, order); err != nil {
			return err
		}
		return s.events.Append(ctx, newEvent(now, domain.EventPaymentCompleted, "payment", p.ID, machineID,
			map[string]any{"order_id": order.ID, "amount_cents": p.AmountCents,
				"discount_cents": p.DiscountCents, "transaction_id": txID}))
	})
	switch {
	case errors.Is(err, errAlreadySettled):
		return nil, fmt.Errorf("%w: payment %s was already settled", domain.ErrConflict, paymentID)
	case err != nil:
		log.Printf("payment %s charged (transaction %s) but not recorded, reconciliation will retry: %v", paymentID, txID, err)
		return nil, &domain.PendingPaymentError{PaymentID: paymentID}
	case unclaimed:
		return nil, s.refundUnclaimed(ctx, payment, machineID)
	}

	if payment.CouponCode != "" {
		s.redeemCoupon(ctx, payment, order, machineID)
	}
	return &domain.Checkout{Order: order, Payment: payment}, nil
}

func (s *vendingService) failPayment(ctx context.Context, paymentID, reason, machineID string) error {
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		p, err := s.payments.GetForUpdate(ctx, paymentID)
		if err != nil {
			return err
		}
		if p.Status != domain.PaymentPending {
			return nil
		}
		now := s.now().UTC()
		p.Status = domain.PaymentFailed
		p.Metadata["error"] = reason
		p.UpdatedAt = now
		if err := s.payments.Update(ctx, p); err != nil {
			return err
		}
		return s.events.Append(ctx, newEvent(now, domain.EventPaymentFailed, "payment", p.ID, machineID,
			map[string]any{"amount_cents": p.AmountCents, "method": p.Method, "error": reason}))
	})
}

func (s *vendingService) redeemCoupon(ctx context.Context, p *domain.Payment, o *domain.Order, machineID string) {
	err := s.coupons.Redeem(ctx, p.CouponCode, o.OrderNo, p.AmountCents+p.DiscountCents)
	if err == nil {
		return
	}
	log.Printf("redeem coupon %s for order %s: %v", p.CouponCode, o.ID, err)
	e := newEvent(s.now().UTC(), domain.EventCouponRedeemFailed, "order", o.ID, machineID,
		map[string]any{"coupon_code": p.CouponCode, "order_no": o.OrderNo, "error": err.Error()})
	if err := s.events.Append(ctx, e); err != nil {
		log.Printf("record coupon failure for order %s: %v", o.ID, err)
	}
}

func (s *vendingService) refund(ctx context.Context, p *domain.Payment, reason string) error {
	if p.AmountCents == 0 || p.TransactionID == "" {
		return nil
	}
	return s.gateway.Refund(ctx, port.RefundRequest{
		PaymentID:     p.ID,
		TransactionID: p.TransactionID,
		AmountCents:   p.AmountCents,
		Method:        p.Method,
		Reason:        reason,
	})
}

func (s *vendingService) refundUnclaimed(ctx context.Context, payment *domain.Payment, machineID string) error {
	if err := s.refund(ctx, payment, "reservation closed before payment completed"); err != nil {
		log.Printf("refund payment %s (transaction %s): %v", payment.ID, payment.TransactionID, err)
		payment.Metadata["refund_error"] = err.Error()
		if err := s.finishPayment(ctx, payment, domain.EventRefundFailed, machineID); err != nil {
			log.Printf("record payment %s: %v", payment.ID, err)
		}
		return fmt.Errorf("%w: charged after the reservation closed and the refund failed", domain.ErrPaymentFailed)
	}
	payment.Status = domain.PaymentRefunded
	if err := s.finishPayment(ctx, payment, domain.EventPaymentRefunded, machineID); err != nil {
		log.Printf("record refunded payment %s: %v", payment.ID, err)
	}
	return domain.ErrReservationClosed
}

func (s *vendingService) finishPayment(ctx context.Context, p *domain.Payment, eventType, machineID string) error {
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		now := s.now().UTC()
		p.UpdatedAt = now
		if p.Status == domain.PaymentCompleted && p.CompletedAt == nil {
			p.CompletedAt = &now
		}
		if err := s.payments.Update(ctx, p); err != nil {
			return err
		}
		data := map[string]any{"amount_cents": p.AmountCents, "transaction_id": p.TransactionID, "method": p.Method}
		if msg, ok := p.Metadata["refund_error"]; ok {
			data["error"] = msg
		}
		return s.events.Append(ctx, newEvent(now, eventType, "payment", p.ID, machineID, data))
	})
}

func (s *vendingService) GetPayment(ctx context.Context, id string) (*domain.Checkout, error) {
	p, err := s.payments.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	out := &domain.Checkout{Payment: p}
	o, err := s.orders.GetByPayment(ctx, id)
	switch {
	case err == nil:
		out.Order = o
	case !errors.Is(err, domain.ErrNotFound):
		return nil, err
	}
	return out, nil
}

func (s *vendingService) ReconcilePayments(ctx context.Context) (int, error) {
	ids, err := s.payments.ListStalePending(ctx, s.now().UTC().Add(-s.cfg.ReconcileAfter), reconcileBatch)
	if err != nil {
		return 0, err
	}
	settled := 0
	var errs []error
	for _, id := range ids {
		if err := s.reconcileOne(ctx, id); err != nil {
			errs = append(errs, fmt.Errorf("payment %s: %w", id, err))
			continue
		}
		settled++
	}
	return settled, errors.Join(errs...)
}

func (s *vendingService) reconcileOne(ctx context.Context, paymentID string) error {
	p, err := s.payments.Get(ctx, paymentID)
	if err != nil {
		return err
	}
	if p.Status != domain.PaymentPending {
		return nil
	}
	sess, err := s.sessions.Get(ctx, p.SessionID)
	if err != nil {
		return err
	}
	txID, found, err := s.gateway.Lookup(ctx, port.LookupRequest{
		PaymentID: p.ID,
		Method:    p.Method,
		WalletID:  p.Metadata[metadataWalletID],
	})
	switch {
	case err != nil:
		return fmt.Errorf("lookup: %w", err)
	case !found && p.AmountCents > 0:
		return s.failPayment(ctx, p.ID, "no charge found at the payment provider", sess.MachineID)
	}
	_, err = s.settle(ctx, p.ID, txID, sess.MachineID)
	var pending *domain.PendingPaymentError
	switch {
	case errors.Is(err, domain.ErrReservationClosed), errors.Is(err, domain.ErrConflict):
		return nil
	case errors.As(err, &pending):
		return fmt.Errorf("settle: still pending")
	}
	return err
}
