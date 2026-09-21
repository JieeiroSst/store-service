package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type fakeOrderRepo struct {
	byID   map[int64]*model.Order
	nextID int64
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{byID: map[int64]*model.Order{}}
}

func (r *fakeOrderRepo) Create(_ context.Context, o *model.Order) (*model.Order, error) {
	r.nextID++
	cp := *o
	cp.ID = r.nextID
	cp.CreatedAt = time.Now()
	cp.UpdatedAt = cp.CreatedAt
	r.byID[cp.ID] = &cp
	return &cp, nil
}

func (r *fakeOrderRepo) GetByID(_ context.Context, id int64) (*model.Order, error) {
	o, ok := r.byID[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *o
	return &cp, nil
}

func (r *fakeOrderRepo) ListByCustomer(_ context.Context, customerID string) ([]model.Order, error) {
	var out []model.Order
	for _, o := range r.byID {
		if o.CustomerID == customerID {
			out = append(out, *o)
		}
	}
	return out, nil
}

func (r *fakeOrderRepo) ListByRestaurant(_ context.Context, restaurantID string) ([]model.Order, error) {
	var out []model.Order
	for _, o := range r.byID {
		if o.RestaurantID == restaurantID {
			out = append(out, *o)
		}
	}
	return out, nil
}

func (r *fakeOrderRepo) ListByDriver(_ context.Context, driverID string) ([]model.Order, error) {
	var out []model.Order
	for _, o := range r.byID {
		if o.DriverID != nil && *o.DriverID == driverID {
			out = append(out, *o)
		}
	}
	return out, nil
}

func (r *fakeOrderRepo) UpdateStatus(_ context.Context, id int64, status model.OrderStatus) error {
	o, ok := r.byID[id]
	if !ok {
		return port.ErrNotFound
	}
	o.Status = status
	return nil
}

func (r *fakeOrderRepo) AssignDriver(_ context.Context, id int64, driverID string) error {
	o, ok := r.byID[id]
	if !ok {
		return port.ErrNotFound
	}
	o.DriverID = &driverID
	return nil
}

func (r *fakeOrderRepo) SetActualDeliveryTime(_ context.Context, id int64, deliveredAt time.Time) error {
	o, ok := r.byID[id]
	if !ok {
		return port.ErrNotFound
	}
	o.ActualDeliveryTime = &deliveredAt
	return nil
}

type fakeTrackingRepo struct {
	events []model.OrderTracking
}

func (r *fakeTrackingRepo) Create(_ context.Context, e *model.OrderTracking) error {
	r.events = append(r.events, *e)
	return nil
}

func (r *fakeTrackingRepo) ListByOrder(_ context.Context, orderID int64) ([]model.OrderTracking, error) {
	var out []model.OrderTracking
	for _, e := range r.events {
		if e.OrderID == orderID {
			out = append(out, e)
		}
	}
	return out, nil
}

type fakeAssignmentRepo struct {
	byID   map[int64]*model.DriverAssignment
	nextID int64
}

func newFakeAssignmentRepo() *fakeAssignmentRepo {
	return &fakeAssignmentRepo{byID: map[int64]*model.DriverAssignment{}}
}

func (r *fakeAssignmentRepo) Create(_ context.Context, a *model.DriverAssignment) (*model.DriverAssignment, error) {
	r.nextID++
	cp := *a
	cp.ID = r.nextID
	r.byID[cp.ID] = &cp
	return &cp, nil
}

func (r *fakeAssignmentRepo) GetByID(_ context.Context, id int64) (*model.DriverAssignment, error) {
	a, ok := r.byID[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (r *fakeAssignmentRepo) Update(_ context.Context, a *model.DriverAssignment) (*model.DriverAssignment, error) {
	r.byID[a.ID] = a
	return a, nil
}

func (r *fakeAssignmentRepo) ListByDriver(_ context.Context, driverID string) ([]model.DriverAssignment, error) {
	var out []model.DriverAssignment
	for _, a := range r.byID {
		if a.DriverID == driverID {
			out = append(out, *a)
		}
	}
	return out, nil
}

type fakeUserClient struct {
	customers map[string]*model.Customer
	drivers   map[string]*model.Driver
}

func newFakeUserClient() *fakeUserClient {
	return &fakeUserClient{customers: map[string]*model.Customer{}, drivers: map[string]*model.Driver{}}
}

func (c *fakeUserClient) GetCustomer(_ context.Context, customerID string) (*model.Customer, error) {
	if cust, ok := c.customers[customerID]; ok {
		return cust, nil
	}
	return nil, port.ErrNotFound
}

func (c *fakeUserClient) GetDriver(_ context.Context, driverID string) (*model.Driver, error) {
	if d, ok := c.drivers[driverID]; ok {
		return d, nil
	}
	return nil, port.ErrNotFound
}

type fakeRestaurantClient struct {
	restaurants map[string]*model.Restaurant
	menuItems   map[string]*model.MenuItem
}

func newFakeRestaurantClient() *fakeRestaurantClient {
	return &fakeRestaurantClient{restaurants: map[string]*model.Restaurant{}, menuItems: map[string]*model.MenuItem{}}
}

func (c *fakeRestaurantClient) GetRestaurant(_ context.Context, restaurantID string) (*model.Restaurant, error) {
	if r, ok := c.restaurants[restaurantID]; ok {
		return r, nil
	}
	return nil, port.ErrNotFound
}

func (c *fakeRestaurantClient) GetMenuItems(_ context.Context, menuItemIDs []string) (map[string]*model.MenuItem, error) {
	out := make(map[string]*model.MenuItem, len(menuItemIDs))
	for _, id := range menuItemIDs {
		if item, ok := c.menuItems[id]; ok {
			out[id] = item
		}
	}
	return out, nil
}

type fakePaymentClient struct {
	authorizeStatus string
	authorizeErr    error
}

func (c *fakePaymentClient) AuthorizePayment(_ context.Context, in port.AuthorizePaymentInput) (*model.PaymentAuthorization, error) {
	if c.authorizeErr != nil {
		return nil, c.authorizeErr
	}
	status := c.authorizeStatus
	if status == "" {
		status = model.PaymentStatusAuthorized
	}
	return &model.PaymentAuthorization{PaymentID: "pay_1", Status: status}, nil
}

type fakeNotifierClient struct {
	notifications []port.NotifyInput
}

func (c *fakeNotifierClient) Notify(_ context.Context, in port.NotifyInput) error {
	c.notifications = append(c.notifications, in)
	return nil
}
