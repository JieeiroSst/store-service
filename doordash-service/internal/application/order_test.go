package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type orderTestDeps struct {
	orders      *fakeOrderRepo
	tracking    *fakeTrackingRepo
	users       *fakeUserClient
	restaurants *fakeRestaurantClient
	payments    *fakePaymentClient
	notifier    *fakeNotifierClient
	svc         port.OrderUsecase
}

func newOrderTestDeps() *orderTestDeps {
	d := &orderTestDeps{
		orders:      newFakeOrderRepo(),
		tracking:    &fakeTrackingRepo{},
		users:       newFakeUserClient(),
		restaurants: newFakeRestaurantClient(),
		payments:    &fakePaymentClient{},
		notifier:    &fakeNotifierClient{},
	}
	d.svc = NewOrderService(d.orders, d.tracking, d.users, d.restaurants, d.payments, d.notifier)
	return d
}

func (d *orderTestDeps) seedHappyPath() {
	d.users.customers["cust-1"] = &model.Customer{ID: "cust-1", Name: "Alice"}
	d.restaurants.restaurants["rest-1"] = &model.Restaurant{ID: "rest-1", Name: "Pizza Place", IsActive: true}
	d.restaurants.menuItems["item-1"] = &model.MenuItem{ID: "item-1", RestaurantID: "rest-1", Name: "Margherita", Price: 12.5, IsActive: true}
}

func baseCreateInput() port.CreateOrderInput {
	return port.CreateOrderInput{
		CustomerID:        "cust-1",
		RestaurantID:      "rest-1",
		DeliveryAddressID: "addr-1",
		PaymentMethodID:   "pm-1",
		DeliveryFee:       2,
		ServiceFee:        1,
		Tax:               1.5,
		Items: []port.CreateOrderItemInput{
			{MenuItemID: "item-1", Quantity: 2},
		},
	}
}

func TestCreateOrder(t *testing.T) {
	t.Run("happy path prices from the restaurant catalog and authorizes payment", func(t *testing.T) {
		d := newOrderTestDeps()
		d.seedHappyPath()

		order, err := d.svc.CreateOrder(context.Background(), baseCreateInput())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Subtotal != 25 {
			t.Errorf("subtotal = %v, want 25 (2 x 12.5)", order.Subtotal)
		}
		if order.TotalAmount != 29.5 {
			t.Errorf("total = %v, want 29.5", order.TotalAmount)
		}
		if order.Status != model.OrderStatusCreated {
			t.Errorf("status = %v, want created", order.Status)
		}
		if order.PaymentStatus != model.PaymentStatusAuthorized {
			t.Errorf("payment status = %v, want authorized", order.PaymentStatus)
		}
		if len(d.tracking.events) != 1 {
			t.Errorf("tracking events = %d, want 1", len(d.tracking.events))
		}
		if len(d.notifier.notifications) != 1 {
			t.Errorf("notifications = %d, want 1", len(d.notifier.notifications))
		}
	})

	t.Run("rejects an empty item list", func(t *testing.T) {
		d := newOrderTestDeps()
		in := baseCreateInput()
		in.Items = nil

		_, err := d.svc.CreateOrder(context.Background(), in)
		if !errors.Is(err, port.ErrEmptyOrder) {
			t.Fatalf("err = %v, want ErrEmptyOrder", err)
		}
	})

	t.Run("rejects an inactive restaurant", func(t *testing.T) {
		d := newOrderTestDeps()
		d.seedHappyPath()
		d.restaurants.restaurants["rest-1"].IsActive = false

		_, err := d.svc.CreateOrder(context.Background(), baseCreateInput())
		if !errors.Is(err, port.ErrRestaurantInactive) {
			t.Fatalf("err = %v, want ErrRestaurantInactive", err)
		}
	})

	t.Run("rejects a menu item the restaurant doesn't have", func(t *testing.T) {
		d := newOrderTestDeps()
		d.seedHappyPath()
		delete(d.restaurants.menuItems, "item-1")

		_, err := d.svc.CreateOrder(context.Background(), baseCreateInput())
		if !errors.Is(err, port.ErrMenuItemUnavailable) {
			t.Fatalf("err = %v, want ErrMenuItemUnavailable", err)
		}
	})

	t.Run("rejects when payment authorization fails", func(t *testing.T) {
		d := newOrderTestDeps()
		d.seedHappyPath()
		d.payments.authorizeStatus = model.PaymentStatusFailed

		_, err := d.svc.CreateOrder(context.Background(), baseCreateInput())
		if !errors.Is(err, port.ErrPaymentAuthorizationFailed) {
			t.Fatalf("err = %v, want ErrPaymentAuthorizationFailed", err)
		}
		if len(d.orders.byID) != 0 {
			t.Error("order should not be persisted when payment authorization fails")
		}
	})
}

func TestUpdateOrderStatus(t *testing.T) {
	t.Run("walks the lifecycle forward", func(t *testing.T) {
		d := newOrderTestDeps()
		d.seedHappyPath()
		order, err := d.svc.CreateOrder(context.Background(), baseCreateInput())
		if err != nil {
			t.Fatalf("create order: %v", err)
		}

		for _, next := range []model.OrderStatus{
			model.OrderStatusConfirmed, model.OrderStatusPreparing,
			model.OrderStatusReadyForPickup, model.OrderStatusPickedUp, model.OrderStatusDelivered,
		} {
			updated, err := d.svc.UpdateOrderStatus(context.Background(), order.ID, next)
			if err != nil {
				t.Fatalf("transition to %s: %v", next, err)
			}
			if updated.Status != next {
				t.Fatalf("status = %v, want %v", updated.Status, next)
			}
		}

		final, _ := d.orders.GetByID(context.Background(), order.ID)
		if final.ActualDeliveryTime == nil {
			t.Error("expected ActualDeliveryTime to be set once delivered")
		}
	})

	t.Run("rejects skipping a step", func(t *testing.T) {
		d := newOrderTestDeps()
		d.seedHappyPath()
		order, _ := d.svc.CreateOrder(context.Background(), baseCreateInput())

		_, err := d.svc.UpdateOrderStatus(context.Background(), order.ID, model.OrderStatusDelivered)
		if !errors.Is(err, port.ErrInvalidStatusTransition) {
			t.Fatalf("err = %v, want ErrInvalidStatusTransition", err)
		}
	})
}

func TestCancelOrder(t *testing.T) {
	t.Run("cancels a freshly created order", func(t *testing.T) {
		d := newOrderTestDeps()
		d.seedHappyPath()
		order, _ := d.svc.CreateOrder(context.Background(), baseCreateInput())

		cancelled, err := d.svc.CancelOrder(context.Background(), order.ID, "changed my mind")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cancelled.Status != model.OrderStatusCancelled {
			t.Errorf("status = %v, want cancelled", cancelled.Status)
		}
	})

	t.Run("rejects cancelling an already terminal order", func(t *testing.T) {
		d := newOrderTestDeps()
		d.seedHappyPath()
		order, _ := d.svc.CreateOrder(context.Background(), baseCreateInput())
		if _, err := d.svc.CancelOrder(context.Background(), order.ID, "first cancel"); err != nil {
			t.Fatalf("first cancel: %v", err)
		}

		_, err := d.svc.CancelOrder(context.Background(), order.ID, "second cancel")
		if !errors.Is(err, port.ErrOrderAlreadyTerminal) {
			t.Fatalf("err = %v, want ErrOrderAlreadyTerminal", err)
		}
	})
}
