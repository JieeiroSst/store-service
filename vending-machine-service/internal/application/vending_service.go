package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/config"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

const sweepBatch = 100

type vendingService struct {
	tx           port.Transactor
	machines     port.MachineRepository
	inventory    port.InventoryRepository
	sessions     port.SessionRepository
	reservations port.ReservationRepository
	payments     port.PaymentRepository
	orders       port.OrderRepository
	events       port.EventRepository
	gateway      port.PaymentGateway
	coupons      port.CouponProvider
	cfg          config.VendingConfig
	now          func() time.Time
}

type VendingDeps struct {
	fx.In

	Tx           port.Transactor
	Machines     port.MachineRepository
	Inventory    port.InventoryRepository
	Sessions     port.SessionRepository
	Reservations port.ReservationRepository
	Payments     port.PaymentRepository
	Orders       port.OrderRepository
	Events       port.EventRepository
	Gateway      port.PaymentGateway
	Coupons      port.CouponProvider
}

func NewVendingService(d VendingDeps, cfg *config.Config) port.VendingService {
	return &vendingService{
		tx:           d.Tx,
		machines:     d.Machines,
		inventory:    d.Inventory,
		sessions:     d.Sessions,
		reservations: d.Reservations,
		payments:     d.Payments,
		orders:       d.Orders,
		events:       d.Events,
		gateway:      d.Gateway,
		coupons:      d.Coupons,
		cfg:          cfg.Vending,
		now:          time.Now,
	}
}

func (s *vendingService) StartSession(ctx context.Context, machineID string) (*domain.Session, error) {
	var sess *domain.Session
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		m, err := s.machines.Get(ctx, machineID)
		if err != nil {
			return err
		}
		if !m.CanSell() {
			return domain.ErrMachineUnavailable
		}
		now := s.now().UTC()
		sess = &domain.Session{
			ID:        uuid.NewString(),
			MachineID: machineID,
			StartedAt: now,
			ExpiresAt: now.Add(s.cfg.SessionTTL),
			Status:    domain.SessionActive,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.sessions.Create(ctx, sess); err != nil {
			return err
		}
		return s.events.Append(ctx, newEvent(now, domain.EventSessionStarted, "session", sess.ID, machineID, nil))
	})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *vendingService) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	return s.sessions.Get(ctx, id)
}

func (s *vendingService) EndSession(ctx context.Context, id string) (*domain.Session, error) {
	var sess *domain.Session
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if sess, err = s.sessions.GetForUpdate(ctx, id); err != nil {
			return err
		}
		return s.closeSession(ctx, sess, domain.SessionCompleted)
	})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *vendingService) closeSession(ctx context.Context, sess *domain.Session, status domain.SessionStatus) error {
	if err := sess.End(status); err != nil {
		return err
	}
	now := s.now().UTC()
	sess.UpdatedAt = now
	if err := s.sessions.Update(ctx, sess); err != nil {
		return err
	}
	pending, err := s.reservations.ListPendingBySession(ctx, sess.ID)
	if err != nil {
		return err
	}
	for _, id := range pending {
		r, err := s.reservations.GetForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if err := s.releaseReservation(ctx, r, domain.ReservationCancelled, sess.MachineID); err != nil {
			return err
		}
	}
	return s.events.Append(ctx, newEvent(now, domain.EventSessionEnded, "session", sess.ID, sess.MachineID,
		map[string]any{"status": status, "released_reservations": len(pending)}))
}

func (s *vendingService) Reserve(ctx context.Context, sessionID, slot string) (*domain.Reservation, error) {
	if slot == "" {
		return nil, fmt.Errorf("%w: slot is required", domain.ErrInvalidInput)
	}
	var r *domain.Reservation
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		sess, err := s.sessions.GetForUpdate(ctx, sessionID)
		if err != nil {
			return err
		}
		now := s.now().UTC()
		if !sess.IsActive(now) {
			return domain.ErrSessionNotActive
		}
		m, err := s.machines.Get(ctx, sess.MachineID)
		if err != nil {
			return err
		}
		if !m.CanSell() {
			return domain.ErrMachineUnavailable
		}
		inv, err := s.inventory.GetBySlotForUpdate(ctx, sess.MachineID, slot)
		if err != nil {
			return err
		}
		if inv.Product == nil || !inv.Product.IsActive {
			return domain.ErrProductInactive
		}
		if err := inv.Take(); err != nil {
			return err
		}
		inv.UpdatedAt = now
		if err := s.inventory.Update(ctx, inv); err != nil {
			return err
		}

		expires := now.Add(s.cfg.ReservationTTL)
		if sess.ExpiresAt.Before(expires) {
			expires = sess.ExpiresAt
		}
		r = &domain.Reservation{
			ID:          uuid.NewString(),
			SessionID:   sess.ID,
			InventoryID: inv.ID,
			ExpiresAt:   expires,
			Status:      domain.ReservationPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.reservations.Create(ctx, r); err != nil {
			return err
		}
		events := []*domain.Event{newEvent(now, domain.EventReservationCreated, "reservation", r.ID, sess.MachineID,
			map[string]any{"slot": inv.SlotIdentifier, "product_id": inv.ProductID})}
		stock := map[string]any{"slot": inv.SlotIdentifier, "product": inv.Product.Name,
			"quantity": inv.Quantity, "low_threshold": inv.LowThreshold, "location": m.Location}
		switch {
		case inv.Quantity == 0:
			events = append(events, newEvent(now, domain.EventInventoryEmpty, "inventory", inv.ID, sess.MachineID, stock))
		case inv.IsLow():
			events = append(events, newEvent(now, domain.EventInventoryLow, "inventory", inv.ID, sess.MachineID, stock))
		}
		return appendEvents(ctx, s.events, events...)
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *vendingService) CancelReservation(ctx context.Context, id string) (*domain.Reservation, error) {
	var r *domain.Reservation
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if r, err = s.reservations.GetForUpdate(ctx, id); err != nil {
			return err
		}
		return s.releaseReservation(ctx, r, domain.ReservationCancelled, "")
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *vendingService) releaseReservation(ctx context.Context, r *domain.Reservation, status domain.ReservationStatus, machineID string) error {
	if err := r.Release(status); err != nil {
		return err
	}
	now := s.now().UTC()
	r.UpdatedAt = now
	if err := s.reservations.Update(ctx, r); err != nil {
		return err
	}
	inv, err := s.inventory.GetForUpdate(ctx, r.InventoryID)
	if err != nil {
		return err
	}
	inv.Return()
	inv.UpdatedAt = now
	if err := s.inventory.Update(ctx, inv); err != nil {
		return err
	}
	if machineID == "" {
		machineID = inv.MachineID
	}
	return s.events.Append(ctx, newEvent(now, domain.EventReservationReleased, "reservation", r.ID, machineID,
		map[string]any{"status": status}))
}

type quote struct {
	priceCents    int
	discountCents int
	machineID     string
}

func (s *vendingService) quote(ctx context.Context, reservationID string, in port.CheckoutInput) (*quote, error) {
	r, err := s.reservations.Get(ctx, reservationID)
	if err != nil {
		return nil, err
	}
	if r.Status != domain.ReservationPending || !s.now().Before(r.ExpiresAt) {
		return nil, domain.ErrReservationClosed
	}
	sess, err := s.sessions.Get(ctx, r.SessionID)
	if err != nil {
		return nil, err
	}
	inv, err := s.inventory.Get(ctx, r.InventoryID)
	if err != nil {
		return nil, err
	}
	if inv.Product == nil {
		return nil, fmt.Errorf("inventory %s has no product", inv.ID)
	}
	q := &quote{priceCents: inv.Product.PriceCents, machineID: sess.MachineID}
	if in.CouponCode != "" {
		discount, err := s.coupons.Quote(ctx, in.CouponCode, q.priceCents)
		if err != nil {
			return nil, err
		}
		q.discountCents = max(0, min(discount, q.priceCents))
	}
	return q, nil
}

func validateCheckout(in *port.CheckoutInput) error {
	in.CouponCode = strings.TrimSpace(in.CouponCode)
	in.WalletID = strings.TrimSpace(in.WalletID)
	if in.Method == domain.PaymentWallet && in.WalletID == "" {
		return fmt.Errorf("%w: wallet_id is required for wallet payments", domain.ErrInvalidInput)
	}
	if in.Method != domain.PaymentWallet && in.WalletID != "" {
		return fmt.Errorf("%w: wallet_id is only allowed for wallet payments", domain.ErrInvalidInput)
	}
	return nil
}

func (s *vendingService) Checkout(ctx context.Context, reservationID string, in port.CheckoutInput) (*domain.Checkout, error) {
	if err := validateCheckout(&in); err != nil {
		return nil, err
	}
	q, err := s.quote(ctx, reservationID, in)
	if err != nil {
		return nil, err
	}

	var payment *domain.Payment
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		r, err := s.reservations.GetForUpdate(ctx, reservationID)
		if err != nil {
			return err
		}
		now := s.now().UTC()
		if r.Status != domain.ReservationPending || !now.Before(r.ExpiresAt) {
			return domain.ErrReservationClosed
		}
		sess, err := s.sessions.Get(ctx, r.SessionID)
		if err != nil {
			return err
		}
		if !sess.IsActive(now) {
			return domain.ErrSessionNotActive
		}
		inv, err := s.inventory.GetForUpdate(ctx, r.InventoryID)
		if err != nil {
			return err
		}
		if inv.Product.PriceCents != q.priceCents {
			return domain.ErrPriceChanged
		}
		metadata := map[string]string{"product_id": inv.ProductID, "slot": inv.SlotIdentifier}
		if in.WalletID != "" {
			metadata["wallet_id"] = in.WalletID
		}
		payment = &domain.Payment{
			ID:            uuid.NewString(),
			SessionID:     sess.ID,
			ReservationID: r.ID,
			AmountCents:   q.priceCents - q.discountCents,
			DiscountCents: q.discountCents,
			CouponCode:    in.CouponCode,
			Currency:      s.cfg.Currency,
			Method:        in.Method,
			Status:        domain.PaymentPending,
			Metadata:      metadata,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		err = s.payments.Create(ctx, payment)
		if errors.Is(err, domain.ErrConflict) {
			return fmt.Errorf("%w: a payment for this reservation is already in progress", domain.ErrConflict)
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	var txID string
	var chargeErr error
	if payment.AmountCents > 0 {
		txID, chargeErr = s.gateway.Charge(ctx, port.ChargeRequest{
			PaymentID:   payment.ID,
			AmountCents: payment.AmountCents,
			Currency:    payment.Currency,
			Method:      payment.Method,
			WalletID:    in.WalletID,
		})
	}
	ctx = context.WithoutCancel(ctx)
	switch {
	case errors.Is(chargeErr, domain.ErrPaymentPending):
		log.Printf("payment %s: charge outcome unknown, left pending for reconciliation: %v", payment.ID, chargeErr)
		return nil, &domain.PendingPaymentError{PaymentID: payment.ID}
	case chargeErr != nil:
		if err := s.failPayment(ctx, payment.ID, chargeErr.Error(), q.machineID); err != nil {
			log.Printf("record failed payment %s: %v", payment.ID, err)
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrPaymentFailed, chargeErr)
	}
	return s.settle(ctx, payment.ID, txID, q.machineID)
}

func (s *vendingService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return s.orders.Get(ctx, id)
}

func (s *vendingService) ListSessionOrders(ctx context.Context, sessionID string) ([]domain.Order, error) {
	if _, err := s.sessions.Get(ctx, sessionID); err != nil {
		return nil, err
	}
	return s.orders.ListBySession(ctx, sessionID)
}

func (s *vendingService) ReportDispense(ctx context.Context, orderID string, dispensed bool) (*domain.Order, error) {
	var (
		order     *domain.Order
		machineID string
	)
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if order, err = s.orders.GetForUpdate(ctx, orderID); err != nil {
			return err
		}
		sess, err := s.sessions.Get(ctx, order.SessionID)
		if err != nil {
			return err
		}
		machineID = sess.MachineID
		if order.Status == domain.OrderFailed && !dispensed {
			return nil
		}
		now := s.now().UTC()
		if err := order.Fulfil(dispensed, now); err != nil {
			return err
		}
		order.UpdatedAt = now
		if err := s.orders.Update(ctx, order); err != nil {
			return err
		}
		eventType := domain.EventOrderCompleted
		if !dispensed {
			eventType = domain.EventOrderFailed
		}
		return s.events.Append(ctx, newEvent(now, eventType, "order", order.ID, machineID,
			map[string]any{"order_no": order.OrderNo}))
	})
	if err != nil || dispensed {
		return order, err
	}
	if err := s.refundOrderPayment(context.WithoutCancel(ctx), order, machineID, "item was not dispensed"); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *vendingService) RefundOrder(ctx context.Context, orderID, reason string) (*domain.Order, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("%w: reason is required", domain.ErrInvalidInput)
	}
	var (
		order     *domain.Order
		machineID string
	)
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if order, err = s.orders.GetForUpdate(ctx, orderID); err != nil {
			return err
		}
		sess, err := s.sessions.Get(ctx, order.SessionID)
		if err != nil {
			return err
		}
		machineID = sess.MachineID
		if order.Status == domain.OrderRefunded {
			return nil
		}
		now := s.now().UTC()
		if err := order.Refund(now); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, order); err != nil {
			return err
		}
		return s.events.Append(ctx, newEvent(now, domain.EventOrderRefunded, "order", order.ID, machineID,
			map[string]any{"reason": reason}))
	})
	if err != nil {
		return nil, err
	}
	if err := s.refundOrderPayment(context.WithoutCancel(ctx), order, machineID, reason); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *vendingService) refundOrderPayment(ctx context.Context, order *domain.Order, machineID, reason string) error {
	payment, err := s.payments.Get(ctx, order.PaymentID)
	if err != nil {
		return err
	}
	if payment.Status != domain.PaymentCompleted {
		return nil
	}
	if err := s.refund(ctx, payment, reason); err != nil {
		e := newEvent(s.now().UTC(), domain.EventRefundFailed, "payment", payment.ID, machineID,
			map[string]any{"order_id": order.ID, "amount_cents": payment.AmountCents, "error": err.Error()})
		if appendErr := s.events.Append(ctx, e); appendErr != nil {
			log.Printf("record refund failure for payment %s: %v", payment.ID, appendErr)
		}
		return fmt.Errorf("%w: refund for order %s: %v", domain.ErrPaymentFailed, order.ID, err)
	}
	payment.Status = domain.PaymentRefunded
	return s.finishPayment(ctx, payment, domain.EventPaymentRefunded, machineID)
}

func (s *vendingService) ExpireStale(ctx context.Context) (int, int, error) {
	now := s.now().UTC()

	reservationIDs, err := s.reservations.ListExpiredPending(ctx, now, sweepBatch)
	if err != nil {
		return 0, 0, err
	}
	reservations := 0
	for _, id := range reservationIDs {
		err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
			r, err := s.reservations.GetForUpdate(ctx, id)
			if err != nil {
				return err
			}
			if r.Status != domain.ReservationPending || now.Before(r.ExpiresAt) {
				return nil
			}
			reservations++
			return s.releaseReservation(ctx, r, domain.ReservationExpired, "")
		})
		if err != nil {
			return reservations, 0, fmt.Errorf("expire reservation %s: %w", id, err)
		}
	}

	sessionIDs, err := s.sessions.ListExpired(ctx, now, sweepBatch)
	if err != nil {
		return reservations, 0, err
	}
	sessions := 0
	for _, id := range sessionIDs {
		err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
			sess, err := s.sessions.GetForUpdate(ctx, id)
			if err != nil {
				return err
			}
			if sess.Status != domain.SessionActive || now.Before(sess.ExpiresAt) {
				return nil
			}
			sessions++
			return s.closeSession(ctx, sess, domain.SessionExpired)
		})
		if err != nil {
			return reservations, sessions, fmt.Errorf("expire session %s: %w", id, err)
		}
	}
	return reservations, sessions, nil
}
