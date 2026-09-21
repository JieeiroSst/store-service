package repository

import (
	"context"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
	"gorm.io/gorm"
)

type orderTrackingRepository struct {
	db *gorm.DB
}

func NewOrderTrackingRepository(db *gorm.DB) port.OrderTrackingRepository {
	return &orderTrackingRepository{db: db}
}

func (r *orderTrackingRepository) Create(ctx context.Context, event *model.OrderTracking) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *orderTrackingRepository) ListByOrder(ctx context.Context, orderID int64) ([]model.OrderTracking, error) {
	var events []model.OrderTracking
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("timestamp asc").Find(&events).Error
	return events, err
}
