package repository

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JIeeiroSst/kitchen-service/internal/domain/port"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) port.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category model.Category) error {
	return r.db.WithContext(ctx).Create(&category).Error
}

func (r *categoryRepository) Find(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
