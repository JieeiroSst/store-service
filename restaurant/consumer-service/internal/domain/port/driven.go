package port

import (
	"context"

	"github.com/JIeeiroSst/consumer-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
)

type ConsumerRepository interface {
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
	Create(ctx context.Context, consumer model.Consumer) error
}

// OrderPublisher fans a placed order out to the kitchen, order-book and
// accounting services over the message bus.
type OrderPublisher interface {
	PublishKitchenCreate(ctx context.Context, order model.PlaceOrder) error
	PublishOrderCreated(ctx context.Context, order model.Order) error
	PublishCartAuthorization(ctx context.Context, cart model.CartAuthorization) error
}
