package repository

import (
	"context"

	"github.com/JIeeiroSst/accounting-service/model"
	"gorm.io/gorm"
)

type AuthCarts interface {
	SaveDelivery(ctx context.Context, delivery model.Delivery) error
	SaveOrder(ctx context.Context, order model.Order) error
}

type authCartRepository struct {
	db *gorm.DB
}

func NewAuthCartRepository(db *gorm.DB) *authCartRepository {
	return &authCartRepository{
		db: db,
	}
}

func (r *authCartRepository) SaveDelivery(ctx context.Context, delivery model.Delivery) error {
	if err := r.db.Create(&delivery).Error; err != nil {
		return err
	}
	return nil
}

func (r *authCartRepository) SaveOrder(ctx context.Context, order model.Order) error {
	if err := r.db.Create(&order).Error; err != nil {
		return err
	}
	return nil
}
