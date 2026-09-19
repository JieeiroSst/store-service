package port

import (
	"context"

	"github.com/JIeeiroSst/delivery-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
)

type DeliveryRepository interface {
	Create(ctx context.Context, delivery model.Delivery) error
	UpdateStatus(ctx context.Context, shipID int, status int) error
	Update(ctx context.Context, shipID int, delivery model.Delivery) error
	FindByActive(ctx context.Context) (*model.Delivery, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}
