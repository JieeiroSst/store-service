package repository

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/model"
	"gorm.io/gorm"
)

type Categories interface {
	Create(ctx context.Context, category model.Category) error
	Find(ctx context.Context) ([]model.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategories(db *gorm.DB) *categoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (r *categoryRepository) Create(ctx context.Context, category model.Category) error {
	if err := r.db.Create(&category).Error; err != nil {
		return err
	}
	return nil
}

func (r *categoryRepository) Find(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}
