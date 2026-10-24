package port

import (
	"context"

	"github.com/JIeeiroSst/delivery-service/internal/domain/model"
)

type DeliveryUsecase interface {
	Create(ctx context.Context, delivery *model.Delivery) error
	UpdateStatus(ctx context.Context, shipID int, status string) error
	FindByActive(ctx context.Context) (*model.Delivery, error)
	FindAll(ctx context.Context, pagination model.Pagination) (model.Pagination, error)
	Update(ctx context.Context, shipID int, delivery *model.Delivery) error
}
