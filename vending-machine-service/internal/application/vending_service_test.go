package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

type fixture struct {
	st      *store
	gateway *fakeGateway
	coupons *fakeCoupons
	svc     *vendingService
	clock   time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	st := newStore()
	f := &fixture{st: st, gateway: &fakeGateway{}, coupons: &fakeCoupons{}, clock: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	f.svc = NewVendingService(VendingDeps{
		Tx:           fakeTx{},
		Machines:     machineRepo{st},
		Inventory:    inventoryRepo{st},
		Sessions:     sessionRepo{st},
		Reservations: reservationRepo{st},
		Payments:     paymentRepo{st},
		Orders:       orderRepo{st},
		Events:       eventRepo{st},
		Gateway:      f.gateway,
		Coupons:      f.coupons,
	}, testCfg).(*vendingService)
	f.svc.now = func() time.Time { return f.clock }

	st.machines["m1"] = domain.Machine{ID: "m1", Status: domain.MachineActive}
	st.products["p1"] = domain.Product{ID: "p1", Name: "Cola", PriceCents: 150, IsActive: true}
	st.inventory["i1"] = domain.Inventory{
		ID: "i1", MachineID: "m1", ProductID: "p1", SlotIdentifier: "A1",
		Quantity: 3, MaxCapacity: 10, LowThreshold: 2,
	}
	return f
}

func (f *fixture) reserve(t *testing.T) (*domain.Session, *domain.Reservation) {
	t.Helper()
	ctx := context.Background()
	sess, err := f.svc.StartSession(ctx, "m1")
	if err != nil {
		t.Fatal(err)
	}
	r, err := f.svc.Reserve(ctx, sess.ID, "A1")
	if err != nil {
		t.Fatal(err)
	}
	return sess, r
}

func pay(m domain.PaymentMethod) port.CheckoutInput { return port.CheckoutInput{Method: m} }

func (f *fixture) quantity() int { return f.st.inventory["i1"].Quantity }

func (f *fixture) hasEvent(eventType string) bool {
	for _, e := range f.st.events {
		if e.EventType == eventType {
			return true
		}
	}
	return false
}

func TestStartSessionRequiresActiveMachine(t *testing.T) {
	f := newFixture(t)
	f.st.machines["m1"] = domain.Machine{ID: "m1", Status: domain.MachineMaintenance}
	if _, err := f.svc.StartSession(context.Background(), "m1"); !errors.Is(err, domain.ErrMachineUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestReserveTakesStockAndFlagsLowInventory(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	if f.quantity() != 2 {
		t.Fatalf("quantity = %d, want 2", f.quantity())
	}
	if r.Status != domain.ReservationPending || !r.ExpiresAt.Equal(f.clock.Add(2*time.Minute)) {
		t.Fatalf("reservation %+v", r)
	}
	if !f.hasEvent(domain.EventInventoryLow) {
		t.Fatal("expected inventory.low event at the threshold")
	}
}

func TestReserveOutOfStock(t *testing.T) {
	f := newFixture(t)
	inv := f.st.inventory["i1"]
	inv.Quantity = 0
	f.st.inventory["i1"] = inv
	sess, _ := f.svc.StartSession(context.Background(), "m1")
	if _, err := f.svc.Reserve(context.Background(), sess.ID, "A1"); !errors.Is(err, domain.ErrOutOfStock) {
		t.Fatalf("got %v", err)
	}
}

func TestReserveRejectsExpiredSession(t *testing.T) {
	f := newFixture(t)
	sess, _ := f.svc.StartSession(context.Background(), "m1")
	f.clock = f.clock.Add(6 * time.Minute)
	if _, err := f.svc.Reserve(context.Background(), sess.ID, "A1"); !errors.Is(err, domain.ErrSessionNotActive) {
		t.Fatalf("got %v", err)
	}
}

func TestReservationNeverOutlivesSession(t *testing.T) {
	f := newFixture(t)
	sess, _ := f.svc.StartSession(context.Background(), "m1")
	f.clock = f.clock.Add(4 * time.Minute)
	r, err := f.svc.Reserve(context.Background(), sess.ID, "A1")
	if err != nil {
		t.Fatal(err)
	}
	if !r.ExpiresAt.Equal(sess.ExpiresAt) {
		t.Fatalf("reservation expires %v, session %v", r.ExpiresAt, sess.ExpiresAt)
	}
}

func TestCancelReservationReturnsStock(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	got, err := f.svc.CancelReservation(context.Background(), r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.ReservationCancelled || f.quantity() != 3 {
		t.Fatalf("status=%s quantity=%d", got.Status, f.quantity())
	}
	if _, err := f.svc.CancelReservation(context.Background(), r.ID); !errors.Is(err, domain.ErrReservationClosed) {
		t.Fatalf("second cancel: %v", err)
	}
}

func TestCheckoutCreatesPaidOrder(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	out, err := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard))
	if err != nil {
		t.Fatal(err)
	}
	if out.Payment.Status != domain.PaymentCompleted || out.Payment.AmountCents != 150 || out.Payment.TransactionID != "txn-1" {
		t.Fatalf("payment %+v", out.Payment)
	}
	if out.Order.Status != domain.OrderProcessing || out.Order.PaymentID != out.Payment.ID {
		t.Fatalf("order %+v", out.Order)
	}
	if f.st.reservations[r.ID].Status != domain.ReservationConfirmed {
		t.Fatalf("reservation %s", f.st.reservations[r.ID].Status)
	}
	if f.st.payments[out.Payment.ID].Status != domain.PaymentCompleted {
		t.Fatal("completed payment not persisted")
	}
}

func TestCheckoutTwiceDoesNotChargeTwice(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	if _, err := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard)); !errors.Is(err, domain.ErrReservationClosed) {
		t.Fatalf("got %v", err)
	}
	if f.gateway.charges != 1 {
		t.Fatalf("charged %d times", f.gateway.charges)
	}
}

func TestCheckoutDeclinedKeepsReservation(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	f.gateway.chargeErr = errors.New("card declined")
	if _, err := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard)); !errors.Is(err, domain.ErrPaymentFailed) {
		t.Fatalf("got %v", err)
	}
	if f.st.reservations[r.ID].Status != domain.ReservationPending {
		t.Fatal("reservation should stay pending so the customer can retry")
	}
	for _, p := range f.st.payments {
		if p.Status != domain.PaymentFailed {
			t.Fatalf("payment status %s", p.Status)
		}
	}
	f.gateway.chargeErr = nil
	if _, err := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCash)); err != nil {
		t.Fatalf("retry: %v", err)
	}
}

func TestCheckoutRefundsWhenReservationExpiresDuringCharge(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	f.gateway.onCharge = func() {
		f.clock = f.clock.Add(3 * time.Minute)
		if _, _, err := f.svc.ExpireStale(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard)); !errors.Is(err, domain.ErrReservationClosed) {
		t.Fatalf("got %v", err)
	}
	if len(f.gateway.refunds) != 1 || len(f.st.orders) != 0 {
		t.Fatalf("refunds=%v orders=%d", f.gateway.refunds, len(f.st.orders))
	}
	for _, p := range f.st.payments {
		if p.Status != domain.PaymentRefunded {
			t.Fatalf("payment status %s", p.Status)
		}
	}
	if f.quantity() != 3 {
		t.Fatalf("stock not returned: %d", f.quantity())
	}
}

func TestReportDispense(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	out, _ := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard))

	o, err := f.svc.ReportDispense(context.Background(), out.Order.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != domain.OrderCompleted || o.FulfilledAt == nil {
		t.Fatalf("order %+v", o)
	}
	if _, err := f.svc.ReportDispense(context.Background(), out.Order.ID, false); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("completed order must not be failed afterwards: %v", err)
	}
	if len(f.gateway.refunds) != 0 {
		t.Fatal("completed order was refunded")
	}
}

func TestFailedDispenseRefundsAndCanRetry(t *testing.T) {
	f := newFixture(t)
	_, r := f.reserve(t)
	out, _ := f.svc.Checkout(context.Background(), r.ID, pay(domain.PaymentCard))

	f.gateway.refundErr = errors.New("processor down")
	if _, err := f.svc.ReportDispense(context.Background(), out.Order.ID, false); !errors.Is(err, domain.ErrPaymentFailed) {
		t.Fatalf("got %v", err)
	}
	if f.st.orders[out.Order.ID].Status != domain.OrderFailed {
		t.Fatal("order should be failed even if the refund did not go through")
	}

	f.gateway.refundErr = nil
	if _, err := f.svc.ReportDispense(context.Background(), out.Order.ID, false); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if f.st.payments[out.Payment.ID].Status != domain.PaymentRefunded || len(f.gateway.refunds) != 1 {
		t.Fatalf("payment=%s refunds=%v", f.st.payments[out.Payment.ID].Status, f.gateway.refunds)
	}
	if _, err := f.svc.ReportDispense(context.Background(), out.Order.ID, false); err != nil || len(f.gateway.refunds) != 1 {
		t.Fatalf("err=%v refunds=%v", err, f.gateway.refunds)
	}
}

func TestEndSessionReleasesPendingReservations(t *testing.T) {
	f := newFixture(t)
	sess, _ := f.reserve(t)
	if _, err := f.svc.Reserve(context.Background(), sess.ID, "A1"); err != nil {
		t.Fatal(err)
	}
	if f.quantity() != 1 {
		t.Fatalf("quantity %d", f.quantity())
	}
	got, err := f.svc.EndSession(context.Background(), sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.SessionCompleted || f.quantity() != 3 {
		t.Fatalf("status=%s quantity=%d", got.Status, f.quantity())
	}
}

func TestExpireStale(t *testing.T) {
	f := newFixture(t)
	f.reserve(t)
	f.clock = f.clock.Add(10 * time.Minute)
	res, sess, err := f.svc.ExpireStale(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res != 1 || sess != 1 || f.quantity() != 3 {
		t.Fatalf("reservations=%d sessions=%d quantity=%d", res, sess, f.quantity())
	}
	for _, s := range f.st.sessions {
		if s.Status != domain.SessionExpired {
			t.Fatalf("session %s", s.Status)
		}
	}
}
