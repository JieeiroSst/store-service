package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"gorm.io/gorm"
)

// orderRepository is read-only: order_order is owned by order-processing-service.
type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) port.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetByID(ctx context.Context, id int) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &order, nil
}

func (r *orderRepository) List(ctx context.Context) ([]model.Order, error) {
	var orders []model.Order
	if err := r.db.WithContext(ctx).Find(&orders).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return orders, nil
}
