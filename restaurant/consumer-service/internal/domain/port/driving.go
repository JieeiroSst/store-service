package port

import (
	"context"

	"github.com/JIeeiroSst/consumer-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
)

type ConsumerUsecase interface {
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
	Create(ctx context.Context, consumer *model.Consumer) error
	PlaceOrder(ctx context.Context, order model.PlaceOrder) error
}
