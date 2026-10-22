package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
)

func placeOrder(t *testing.T, d *orderTestDeps) *model.Order {
	t.Helper()
	d.seedHappyPath()
	order, err := d.svc.CreateOrder(context.Background(), baseCreateInput())
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.PaymentID != "pay_1" || order.PaymentStatus != model.PaymentStatusAuthorized {
		t.Fatalf("authorization not recorded on order: %+v", order)
	}
	return order
}

func advance(t *testing.T, d *orderTestDeps, id int64, statuses ...model.OrderStatus) {
	t.Helper()
	for _, s := range statuses {
		if _, err := d.svc.UpdateOrderStatus(context.Background(), id, s); err != nil {
			t.Fatalf("move to %s: %v", s, err)
		}
	}
}

func TestDeliveredOrderCapturesPayment(t *testing.T) {
	d := newOrderTestDeps()
	order := placeOrder(t, d)

	advance(t, d, order.ID, model.OrderStatusConfirmed, model.OrderStatusPreparing, model.OrderStatusReadyForPickup, model.OrderStatusPickedUp)
	if len(d.payments.captured) != 0 {
		t.Fatal("payment captured before delivery")
	}
	advance(t, d, order.ID, model.OrderStatusDelivered)

	if len(d.payments.captured) != 1 || d.payments.captured[0] != "pay_1" {
		t.Fatalf("expected one capture of pay_1, got %v", d.payments.captured)
	}
	if got := d.orders.byID[order.ID].PaymentStatus; got != model.PaymentStatusCaptured {
		t.Fatalf("payment status = %s, want captured", got)
	}
}

func TestCancelledOrderVoidsAuthorization(t *testing.T) {
	d := newOrderTestDeps()
	order := placeOrder(t, d)

	if _, err := d.svc.CancelOrder(context.Background(), order.ID, "changed my mind"); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if len(d.payments.voided) != 1 || d.payments.voided[0] != "pay_1" {
		t.Fatalf("expected one void of pay_1, got %v", d.payments.voided)
	}
	if got := d.orders.byID[order.ID].PaymentStatus; got != model.PaymentStatusVoided {
		t.Fatalf("payment status = %s, want voided", got)
	}
}

func TestSettlementFailureKeepsOrderTransition(t *testing.T) {
	d := newOrderTestDeps()
	order := placeOrder(t, d)
	d.payments.settleErr = errors.New("wallet down")

	cancelled, err := d.svc.CancelOrder(context.Background(), order.ID, "x")
	if err != nil {
		t.Fatalf("cancel must not fail when void fails: %v", err)
	}
	if cancelled.Status != model.OrderStatusCancelled {
		t.Fatalf("status = %s", cancelled.Status)
	}
	if got := d.orders.byID[order.ID].PaymentStatus; got != model.PaymentStatusAuthorized {
		t.Fatalf("payment status must stay authorized for retry, got %s", got)
	}
}

func TestStatusUpdateToCancelledVoidsAuthorization(t *testing.T) {
	d := newOrderTestDeps()
	order := placeOrder(t, d)

	advance(t, d, order.ID, model.OrderStatusConfirmed, model.OrderStatusCancelled)

	if len(d.payments.voided) != 1 {
		t.Fatalf("expected a void, got %v", d.payments.voided)
	}
}
