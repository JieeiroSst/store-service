package repository

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/model"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type Kitchens interface {
	Create(ctx context.Context, kitchen model.Kitchen) error
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

type kitchenRepository struct {
	db *gorm.DB
}

func NewKitchenRepository(db *gorm.DB) *kitchenRepository {
	return &kitchenRepository{
		db: db,
	}
}

func (r *kitchenRepository) Create(ctx context.Context, kitchen model.Kitchen) error {
	if err := r.db.Create(&kitchen).Error; err != nil {
		return err
	}
	return nil
}

func (r *kitchenRepository) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var kitchens []*model.Kitchen

	r.db.Scopes(logger.Paginate(kitchens, &pagination, r.db)).Find(&kitchens)
	pagination.Rows = kitchens

	return pagination, nil
}
