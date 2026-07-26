package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"gorm.io/gorm"
)

type basketLineRepository struct {
	db *gorm.DB
}

func NewBasketLineRepository(db *gorm.DB) port.BasketLineRepository {
	return &basketLineRepository{db: db}
}

func (r *basketLineRepository) Create(ctx context.Context, line *model.BasketLine) error {
	if err := r.db.WithContext(ctx).Create(line).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketLineRepository) GetByID(ctx context.Context, id int) (*model.BasketLine, error) {
	var line model.BasketLine
	if err := r.db.WithContext(ctx).First(&line, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &line, nil
}

func (r *basketLineRepository) Update(ctx context.Context, line *model.BasketLine) error {
	if err := r.db.WithContext(ctx).Model(&model.BasketLine{}).Where("id = ?", line.ID).Updates(line).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketLineRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.BasketLine{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketLineRepository) List(ctx context.Context) ([]model.BasketLine, error) {
	var lines []model.BasketLine
	if err := r.db.WithContext(ctx).Find(&lines).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return lines, nil
}
