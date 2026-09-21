package port

import (
	"context"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
)

type CreateOrderItemInput struct {
	MenuItemID          string
	Quantity            int
	Customizations      string
	SpecialInstructions string
}

type CreateOrderInput struct {
	CustomerID          string
	RestaurantID        string
	DeliveryAddressID   string
	PaymentMethodID     string
	DeliveryFee         float64
	ServiceFee          float64
	Tax                 float64
	Tip                 float64
	SpecialInstructions string
	Items               []CreateOrderItemInput
}

type OrderUsecase interface {
	CreateOrder(ctx context.Context, in CreateOrderInput) (*model.Order, error)
	GetOrder(ctx context.Context, id int64) (*model.Order, error)
	ListOrdersByCustomer(ctx context.Context, customerID string) ([]model.Order, error)
	ListOrdersByRestaurant(ctx context.Context, restaurantID string) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error)
	CancelOrder(ctx context.Context, id int64, reason string) (*model.Order, error)
}

type AddTrackingEventInput struct {
	OrderID   int64
	Status    model.OrderStatus
	Latitude  *float64
	Longitude *float64
	Notes     string
}

type TrackingUsecase interface {
	AddTrackingEvent(ctx context.Context, in AddTrackingEventInput) (*model.OrderTracking, error)
	ListTrackingByOrder(ctx context.Context, orderID int64) ([]model.OrderTracking, error)
}

type DriverAssignmentUsecase interface {
	AssignDriver(ctx context.Context, orderID int64, driverID string) (*model.DriverAssignment, error)
	AcceptAssignment(ctx context.Context, assignmentID int64) (*model.DriverAssignment, error)
	RejectAssignment(ctx context.Context, assignmentID int64, reason string) (*model.DriverAssignment, error)
	CompleteAssignment(ctx context.Context, assignmentID int64) (*model.DriverAssignment, error)
	ListAssignmentsByDriver(ctx context.Context, driverID string) ([]model.DriverAssignment, error)
}
