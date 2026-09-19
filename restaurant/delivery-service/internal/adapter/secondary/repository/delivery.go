package repository

import (
	"context"

	"github.com/JIeeiroSst/delivery-service/internal/domain/model"
	"github.com/JIeeiroSst/delivery-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type deliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) port.DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) Create(ctx context.Context, delivery model.Delivery) error {
	return r.db.WithContext(ctx).Create(&delivery).Error
}

func (r *deliveryRepository) UpdateStatus(ctx context.Context, shipID int, status int) error {
	return r.db.WithContext(ctx).Model(&model.Delivery{}).Where("ship_id = ?", shipID).Update("status", status).Error
}

func (r *deliveryRepository) FindByActive(ctx context.Context) (*model.Delivery, error) {
	var delivery model.Delivery
	if err := r.db.WithContext(ctx).Where("status = ?", model.StatusFree).First(&delivery).Error; err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *deliveryRepository) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var deliveries []*model.Delivery

	r.db.WithContext(ctx).Scopes(logger.Paginate(deliveries, &pagination, r.db)).Find(&deliveries)
	pagination.Rows = deliveries

	return pagination, nil
}

func (r *deliveryRepository) Update(ctx context.Context, shipID int, delivery model.Delivery) error {
	return r.db.WithContext(ctx).Where("ship_id = ?", shipID).Updates(delivery).Error
}
