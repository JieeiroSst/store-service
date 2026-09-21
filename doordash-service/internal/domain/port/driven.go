package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) (*model.Order, error)
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	ListByCustomer(ctx context.Context, customerID string) ([]model.Order, error)
	ListByRestaurant(ctx context.Context, restaurantID string) ([]model.Order, error)
	ListByDriver(ctx context.Context, driverID string) ([]model.Order, error)
	UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) error
	AssignDriver(ctx context.Context, id int64, driverID string) error
	SetActualDeliveryTime(ctx context.Context, id int64, deliveredAt time.Time) error
}

type OrderTrackingRepository interface {
	Create(ctx context.Context, event *model.OrderTracking) error
	ListByOrder(ctx context.Context, orderID int64) ([]model.OrderTracking, error)
}

type DriverAssignmentRepository interface {
	Create(ctx context.Context, assignment *model.DriverAssignment) (*model.DriverAssignment, error)
	GetByID(ctx context.Context, id int64) (*model.DriverAssignment, error)
	Update(ctx context.Context, assignment *model.DriverAssignment) (*model.DriverAssignment, error)
	ListByDriver(ctx context.Context, driverID string) ([]model.DriverAssignment, error)
}

type UserClient interface {
	GetCustomer(ctx context.Context, customerID string) (*model.Customer, error)
	GetDriver(ctx context.Context, driverID string) (*model.Driver, error)
}

type RestaurantClient interface {
	GetRestaurant(ctx context.Context, restaurantID string) (*model.Restaurant, error)
	GetMenuItems(ctx context.Context, menuItemIDs []string) (map[string]*model.MenuItem, error)
}

type AuthorizePaymentInput struct {
	OrderID         int64
	CustomerID      string
	Amount          float64
	PaymentMethodID string
}

type PaymentClient interface {
	AuthorizePayment(ctx context.Context, in AuthorizePaymentInput) (*model.PaymentAuthorization, error)
}

type NotifyInput struct {
	UserID      string
	Title       string
	Message     string
	Type        string
	ReferenceID string
}

type NotifierClient interface {
	Notify(ctx context.Context, in NotifyInput) error
}
