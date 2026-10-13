package http

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type fakeOrders struct {
	port.OrderUsecase
	createdFor string
}

func driver(id string) *string { return &id }

func (f *fakeOrders) GetOrder(_ context.Context, id int64) (*model.Order, error) {
	switch id {
	case 1:
		return &model.Order{ID: 1, CustomerID: "7", DriverID: driver("9")}, nil
	case 2:
		return &model.Order{ID: 2, CustomerID: "8"}, nil
	}
	return nil, port.ErrNotFound
}
func (f *fakeOrders) CreateOrder(_ context.Context, in port.CreateOrderInput) (*model.Order, error) {
	f.createdFor = in.CustomerID
	return &model.Order{CustomerID: in.CustomerID}, nil
}
func (f *fakeOrders) CancelOrder(_ context.Context, id int64, reason string) (*model.Order, error) {
	return &model.Order{ID: id}, nil
}
func (f *fakeOrders) ListOrdersByCustomer(context.Context, string) ([]model.Order, error) {
	return nil, nil
}

type fakeTracking struct{ port.TrackingUsecase }

func (fakeTracking) ListTrackingByOrder(context.Context, int64) ([]model.OrderTracking, error) {
	return nil, nil
}

type fakeAssignments struct{ port.DriverAssignmentUsecase }

func (fakeAssignments) GetAssignment(_ context.Context, id int64) (*model.DriverAssignment, error) {
	return &model.DriverAssignment{ID: id, DriverID: "9"}, nil
}
func (fakeAssignments) AcceptAssignment(_ context.Context, id int64) (*model.DriverAssignment, error) {
	return &model.DriverAssignment{ID: id}, nil
}

func TestDoordashAuthorization(t *testing.T) {
	fakeUserService(t)
	orders := &fakeOrders{}
	router := NewRouter(NewHandler(orders, fakeTracking{}, fakeAssignments{}))
	customer := testToken("7", "user")
	driverToken := testToken("9", "user")
	for _, c := range []struct {
		name string
		call authCall
		want int
	}{
		{"anonymous", authCall{method: "GET", path: "/api/v1/orders/1"}, 401},
		{"own order", authCall{method: "GET", path: "/api/v1/orders/1", bearer: customer}, 200},
		{"foreign order", authCall{method: "GET", path: "/api/v1/orders/2", bearer: customer}, 403},
		{"driver sees assigned order", authCall{method: "GET", path: "/api/v1/orders/1", bearer: driverToken}, 200},
		{"driver cannot cancel", authCall{method: "POST", path: "/api/v1/orders/1/cancel", body: `{}`, bearer: driverToken}, 403},
		{"customer cancels own", authCall{method: "POST", path: "/api/v1/orders/1/cancel", body: `{}`, bearer: customer}, 200},
		{"foreign cancel", authCall{method: "POST", path: "/api/v1/orders/2/cancel", body: `{}`, bearer: customer}, 403},
		{"customer sets status", authCall{method: "PATCH", path: "/api/v1/orders/1/status", body: `{"status":"x"}`, bearer: customer}, 403},
		{"restaurant listing", authCall{method: "GET", path: "/api/v1/orders?restaurant_id=r1", bearer: customer}, 403},
		{"foreign listing", authCall{method: "GET", path: "/api/v1/orders?customer_id=8", bearer: customer}, 403},
		{"own listing", authCall{method: "GET", path: "/api/v1/orders?customer_id=7", bearer: customer}, 200},
		{"own tracking", authCall{method: "GET", path: "/api/v1/orders/1/tracking", bearer: customer}, 200},
		{"customer accepts assignment", authCall{method: "POST", path: "/api/v1/driver-assignments/5/accept", bearer: customer}, 403},
		{"driver accepts assignment", authCall{method: "POST", path: "/api/v1/driver-assignments/5/accept", bearer: driverToken}, 200},
	} {
		if got, body := doAuth(router, c.call); got != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, got, c.want, body)
		}
	}
	doAuth(router, authCall{method: "POST", path: "/api/v1/orders", body: `{"customer_id":"8","restaurant_id":"r","delivery_address_id":"a","payment_method_id":"p","items":[{"menu_item_id":"m","quantity":1}]}`, bearer: customer})
	if orders.createdFor != "7" {
		t.Fatalf("create forced customer: %q", orders.createdFor)
	}
}
