package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/bookStore-service/common"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) port.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *model.Category) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id int) (*model.Category, error) {
	var category model.Category
	if err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &category, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *model.Category) error {
	if err := r.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", category.ID).Updates(category).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Category{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *categoryRepository) List(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return categories, nil
}
