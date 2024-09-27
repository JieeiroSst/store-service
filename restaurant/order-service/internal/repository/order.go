package repository

import (
	"context"

	"github.com/JIeeiroSst/order-service/internal/model"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type Orders interface {
	Create(ctx context.Context, order model.Order) error
	Update(ctx context.Context, id int, order model.Order) error
	FindByID(ctx context.Context, id int) (*model.Order, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *orderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) Create(ctx context.Context, order model.Order) error {
	if err := r.db.Create(order).Error; err != nil {
		return err
	}
	return nil
}

func (r *orderRepository) Update(ctx context.Context, id int, order model.Order) error {
	err := r.db.Model(model.Order{}).Where("id = ? ", id).Updates(order).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *orderRepository) FindByID(ctx context.Context, id int) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Roles").Where("id = ?", id).Find(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var orders []*model.Order

	r.db.Scopes(logger.Paginate(orders, &pagination, r.db)).Find(&orders)
	pagination.Rows = orders

	return pagination, nil
}
