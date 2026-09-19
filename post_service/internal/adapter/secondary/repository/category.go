package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) port.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// Update doesn't turn "0 rows affected" into ErrNotFound - see
// postRepository.Update's doc comment for why (a no-op edit isn't a
// not-found case).
func (r *categoryRepository) Update(ctx context.Context, id string, category *model.Category) error {
	return r.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        category.Name,
			"description": category.Description,
		}).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Category{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) List(ctx context.Context, cursor string, limit int) ([]model.Category, string, error) {
	limit = model.ClampLimit(limit)
	tx := applyDescCursor(r.db.WithContext(ctx), cursor, "created_at", "id")

	var categories []model.Category
	if err := tx.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&categories).Error; err != nil {
		return nil, "", err
	}
	page, next := model.PageFromRows(categories, limit)
	return page, next, nil
}
