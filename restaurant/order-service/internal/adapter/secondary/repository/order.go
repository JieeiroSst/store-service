package repository

import (
	"context"

	"github.com/JIeeiroSst/order-service/internal/domain/model"
	"github.com/JIeeiroSst/order-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) port.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) Update(ctx context.Context, id int, order *model.Order) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", id).Update("status", order.Status).Error
}

func (r *orderRepository) FindByID(ctx context.Context, id int) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", id).Find(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var orders []*model.Order

	r.db.WithContext(ctx).Preload("Items").Scopes(logger.Paginate(orders, &pagination, r.db)).Find(&orders)
	pagination.Rows = orders

	return pagination, nil
}
