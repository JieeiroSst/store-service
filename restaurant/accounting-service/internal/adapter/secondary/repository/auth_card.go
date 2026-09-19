package repository

import (
	"context"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
	"github.com/JIeeiroSst/accounting-service/internal/domain/port"
	"gorm.io/gorm"
)

type authCartRepository struct {
	db *gorm.DB
}

func NewAuthCartRepository(db *gorm.DB) port.AuthCartRepository {
	return &authCartRepository{db: db}
}

func (r *authCartRepository) SaveDelivery(ctx context.Context, delivery model.Delivery) error {
	return r.db.WithContext(ctx).Create(&delivery).Error
}

func (r *authCartRepository) SaveOrder(ctx context.Context, order model.Order) error {
	return r.db.WithContext(ctx).Create(&order).Error
}
