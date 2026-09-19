package port

import (
	"context"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
)

type AuthCartRepository interface {
	SaveDelivery(ctx context.Context, delivery model.Delivery) error
	SaveOrder(ctx context.Context, order model.Order) error
}

type PaymentRepository interface {
	Create(ctx context.Context, payment model.Payment) error
	UpdateStatus(ctx context.Context, orderID int, status string) error
	FindByOrderID(ctx context.Context, orderID int) (*model.Payment, error)
}

// OrderPublisher fans an authorized cart out to the order-book and
// delivery-dispatch services over the message bus.
type OrderPublisher interface {
	PublishOrderSuccess(ctx context.Context, order model.Order) error
	PublishOrderReject(ctx context.Context, order model.Order) error
	PublishDeliveryShip(ctx context.Context, delivery model.Delivery) error
}
