package repository

import (
	"context"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, orderID string) (*entity.Order, error)
	UpdateStatus(ctx context.Context, orderID, status string) error
	GetStaleOrders(ctx context.Context, olderThanMinutes int) ([]entity.StaleOrder, error)
	Delete(ctx context.Context, orderID string) error
}
