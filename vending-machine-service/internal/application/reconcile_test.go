package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

var errTimeout = fmt.Errorf("%w: wallet withdraw: timeout", domain.ErrPaymentPending)

func walletPay() port.CheckoutInput {
	return port.CheckoutInput{Method: domain.PaymentWallet, WalletID: "w1"}
}

func (f *fixture) pendingCheckout(t *testing.T) (*domain.Reservation, string) {
	t.Helper()
	_, r := f.reserve(t)
	f.gateway.chargeErr = errTimeout
	_, err := f.svc.Checkout(context.Background(), r.ID, walletPay())
	var pending *domain.PendingPaymentError
	if !errors.As(err, &pending) || !errors.Is(err, domain.ErrPaymentPending) {
		t.Fatalf("got %v", err)
	}
	f.gateway.chargeErr = nil
	return r, pending.PaymentID
}

func TestUnknownChargeOutcomeStaysPendingAndBlocksRetry(t *testing.T) {
	f := newFixture(t)
	r, paymentID := f.pendingCheckout(t)

	if p := f.st.payments[paymentID]; p.Status != domain.PaymentPending {
		t.Fatalf("payment %s, must stay pending until the provider is asked", p.Status)
	}
	_, err := f.svc.Checkout(context.Background(), r.ID, walletPay())
	if !errors.Is(err, domain.ErrConflict) || f.gateway.charges != 1 {
		t.Fatalf("retry must not charge again: err=%v charges=%d", err, f.gateway.charges)
	}
}

func TestReconcileSettlesChargeFoundAtProvider(t *testing.T) {
	f := newFixture(t)
	_, paymentID := f.pendingCheckout(t)

	f.gateway.lookupTx = "wtx-1"
	if n, err := f.svc.ReconcilePayments(context.Background()); err != nil || n != 0 || f.gateway.lookups != 0 {
		t.Fatalf("must wait for the grace period: n=%d err=%v lookups=%d", n, err, f.gateway.lookups)
	}
	f.clock = f.clock.Add(time.Minute)
	n, err := f.svc.ReconcilePayments(context.Background())
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	out, err := f.svc.GetPayment(context.Background(), paymentID)
	if err != nil {
		t.Fatal(err)
	}
	if out.Payment.Status != domain.PaymentCompleted || out.Payment.TransactionID != "wtx-1" || out.Order == nil {
		t.Fatalf("payment %+v order %+v", out.Payment, out.Order)
	}
}

func TestReconcileFailsPaymentNotFoundAtProvider(t *testing.T) {
	f := newFixture(t)
	r, paymentID := f.pendingCheckout(t)
	f.clock = f.clock.Add(time.Minute)

	if n, err := f.svc.ReconcilePayments(context.Background()); err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if p := f.st.payments[paymentID]; p.Status != domain.PaymentFailed {
		t.Fatalf("payment %s", p.Status)
	}
	if _, err := f.svc.Checkout(context.Background(), r.ID, walletPay()); err != nil {
		t.Fatalf("customer can pay again once the first attempt is known to have failed: %v", err)
	}
}

func TestReconcileKeepsPendingWhenProviderUnreachable(t *testing.T) {
	f := newFixture(t)
	_, paymentID := f.pendingCheckout(t)
	f.clock = f.clock.Add(time.Minute)
	f.gateway.lookupErr = errors.New("wallet service down")

	if _, err := f.svc.ReconcilePayments(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	if p := f.st.payments[paymentID]; p.Status != domain.PaymentPending {
		t.Fatalf("payment %s", p.Status)
	}
}

func TestReconcileRefundsChargeForExpiredReservation(t *testing.T) {
	f := newFixture(t)
	_, paymentID := f.pendingCheckout(t)
	f.clock = f.clock.Add(3 * time.Minute)
	if _, _, err := f.svc.ExpireStale(context.Background()); err != nil {
		t.Fatal(err)
	}

	f.gateway.lookupTx = "wtx-1"
	if n, err := f.svc.ReconcilePayments(context.Background()); err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if p := f.st.payments[paymentID]; p.Status != domain.PaymentRefunded || len(f.gateway.refunds) != 1 {
		t.Fatalf("payment %s refunds %v", p.Status, f.gateway.refunds)
	}
	if len(f.st.orders) != 0 || f.quantity() != 3 {
		t.Fatalf("orders=%d quantity=%d", len(f.st.orders), f.quantity())
	}
}

func TestLateChargeOnFailedPaymentIsRefunded(t *testing.T) {
	f := newFixture(t)
	_, paymentID := f.pendingCheckout(t)
	f.clock = f.clock.Add(time.Minute)
	f.svc.ReconcilePayments(context.Background())

	_, err := f.svc.settle(context.Background(), paymentID, "wtx-late", "m1")
	if !errors.Is(err, domain.ErrReservationClosed) {
		t.Fatalf("got %v", err)
	}
	if p := f.st.payments[paymentID]; p.Status != domain.PaymentRefunded || f.gateway.refunds[0] != "wtx-late" {
		t.Fatalf("payment %s refunds %v", p.Status, f.gateway.refunds)
	}
}

func TestSettleTwiceDoesNothingSecondTime(t *testing.T) {
	f := newFixture(t)
	_, paymentID := f.pendingCheckout(t)
	if _, err := f.svc.settle(context.Background(), paymentID, "wtx-1", "m1"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.settle(context.Background(), paymentID, "wtx-1", "m1"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("got %v", err)
	}
	if len(f.st.orders) != 1 || len(f.gateway.refunds) != 0 {
		t.Fatalf("orders=%d refunds=%v", len(f.st.orders), f.gateway.refunds)
	}
}
