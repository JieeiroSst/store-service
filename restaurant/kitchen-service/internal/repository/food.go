package repository

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/model"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type Foods interface {
	Create(ctx context.Context, food model.Food) error
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

type foodRepository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) *foodRepository {
	return &foodRepository{
		db: db,
	}
}

func (r *foodRepository) Create(ctx context.Context, food model.Food) error {
	if err := r.db.Create(&food).Error; err != nil {
		return err
	}
	return nil
}

func (r *foodRepository) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var foods []*model.Food

	r.db.Scopes(logger.Paginate(foods, &pagination, r.db)).Find(&foods)
	pagination.Rows = foods

	return pagination, nil
}
