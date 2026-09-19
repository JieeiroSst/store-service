package repository

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JIeeiroSst/kitchen-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type foodRepository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) port.FoodRepository {
	return &foodRepository{db: db}
}

func (r *foodRepository) Create(ctx context.Context, food model.Food) error {
	return r.db.WithContext(ctx).Create(&food).Error
}

func (r *foodRepository) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var foods []*model.Food

	r.db.WithContext(ctx).Scopes(logger.Paginate(foods, &pagination, r.db)).Find(&foods)
	pagination.Rows = foods

	return pagination, nil
}

func (r *foodRepository) FindByIDs(ctx context.Context, ids []int) ([]model.Food, error) {
	var foods []model.Food
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&foods).Error; err != nil {
		return nil, err
	}
	return foods, nil
}
