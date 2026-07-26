package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
	"gorm.io/gorm"
)

type adCategoryRepository struct {
	db *gorm.DB
}

func NewAdCategoryRepository(db *gorm.DB) port.AdCategoryRepository {
	return &adCategoryRepository{db: db}
}

func (r *adCategoryRepository) Create(ctx context.Context, category *model.AdCategory) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adCategoryRepository) GetByID(ctx context.Context, id uint) (*model.AdCategory, error) {
	var category model.AdCategory
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &category, nil
}

func (r *adCategoryRepository) Update(ctx context.Context, category *model.AdCategory) error {
	if err := r.db.WithContext(ctx).Model(&model.AdCategory{}).
		Where("id = ?", category.ID).
		Updates(category).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adCategoryRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.AdCategory{}, id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adCategoryRepository) List(ctx context.Context) ([]model.AdCategory, error) {
	var categories []model.AdCategory
	if err := r.db.WithContext(ctx).Order("id desc").Find(&categories).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return categories, nil
}
