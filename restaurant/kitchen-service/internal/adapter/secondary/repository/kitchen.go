package repository

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JIeeiroSst/kitchen-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type kitchenRepository struct {
	db *gorm.DB
}

func NewKitchenRepository(db *gorm.DB) port.KitchenRepository {
	return &kitchenRepository{db: db}
}

func (r *kitchenRepository) Create(ctx context.Context, kitchen model.Kitchen) error {
	return r.db.WithContext(ctx).Create(&kitchen).Error
}

func (r *kitchenRepository) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var kitchens []*model.Kitchen

	r.db.WithContext(ctx).Scopes(logger.Paginate(kitchens, &pagination, r.db)).Find(&kitchens)
	pagination.Rows = kitchens

	return pagination, nil
}

func (r *kitchenRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).Model(&model.Kitchen{}).Where("id = ?", id).Update("status", status).Error
}
