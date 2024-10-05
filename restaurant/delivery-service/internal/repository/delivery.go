package repository

import (
	"context"

	"github.com/JIeeiroSst/delivery-service/internal/model"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type Deliveries interface {
	Create(ctx context.Context, delivery model.Delivery) error
	UpdateStatus(ctx context.Context, shipID int, status int) error
	Update(ctx context.Context, shipID int, delivery model.Delivery) error
	FindByActive(ctx context.Context) (*model.Delivery, error)
	FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

type DeliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository {
	return &DeliveryRepository{
		db: db,
	}
}

func (r *DeliveryRepository) Create(ctx context.Context, delivery model.Delivery) error {
	if err := r.db.Create(&delivery).Error; err != nil {
		return err
	}
	return nil
}

func (r *DeliveryRepository) UpdateStatus(ctx context.Context, shipID int, status int) error {
	err := r.db.Model(model.Delivery{}).Where("active = ?", true).Update("status", status).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *DeliveryRepository) FindByActive(ctx context.Context) (*model.Delivery, error) {
	var delivery model.Delivery
	err := r.db.Where("status = ?", 1).First(&delivery).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *DeliveryRepository) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var deliveries []*model.Delivery

	r.db.Scopes(logger.Paginate(deliveries, &pagination, r.db)).Find(&deliveries)
	pagination.Rows = deliveries

	return pagination, nil
}

func (r *DeliveryRepository) Update(ctx context.Context, shipID int, delivery model.Delivery) error {
	query := r.db.Where("id = ?", shipID).Updates(delivery)
	if query.Error != nil {
		return query.Error
	}
	return nil
}
