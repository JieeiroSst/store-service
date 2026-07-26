package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"gorm.io/gorm"
)

type basketRepository struct {
	db *gorm.DB
}

func NewBasketRepository(db *gorm.DB) port.BasketRepository {
	return &basketRepository{db: db}
}

func (r *basketRepository) Create(ctx context.Context, basket *model.Basket) error {
	if err := r.db.WithContext(ctx).Create(basket).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketRepository) GetByID(ctx context.Context, id int) (*model.Basket, error) {
	var basket model.Basket
	if err := r.db.WithContext(ctx).First(&basket, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &basket, nil
}

func (r *basketRepository) Update(ctx context.Context, basket *model.Basket) error {
	if err := r.db.WithContext(ctx).Model(&model.Basket{}).Where("id = ?", basket.ID).Updates(basket).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Basket{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketRepository) List(ctx context.Context) ([]model.Basket, error) {
	var baskets []model.Basket
	if err := r.db.WithContext(ctx).Find(&baskets).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return baskets, nil
}
