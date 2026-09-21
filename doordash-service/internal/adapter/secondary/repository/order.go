package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) port.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) (*model.Order, error) {
	if err := r.db.WithContext(ctx).Create(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Preload("Items").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) ListByCustomer(ctx context.Context, customerID string) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("placed_at desc").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) ListByRestaurant(ctx context.Context, restaurantID string) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).Where("restaurant_id = ?", restaurantID).Order("placed_at desc").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) ListByDriver(ctx context.Context, driverID string) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).Order("placed_at desc").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) error {
	result := r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

func (r *orderRepository) AssignDriver(ctx context.Context, id int64, driverID string) error {
	result := r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", id).
		Updates(map[string]any{"driver_id": driverID, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

func (r *orderRepository) SetActualDeliveryTime(ctx context.Context, id int64, deliveredAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", id).
		Updates(map[string]any{"actual_delivery_time": deliveredAt, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}
