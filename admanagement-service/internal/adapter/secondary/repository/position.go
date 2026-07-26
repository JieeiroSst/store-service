package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
	"gorm.io/gorm"
)

type adPositionRepository struct {
	db *gorm.DB
}

func NewAdPositionRepository(db *gorm.DB) port.AdPositionRepository {
	return &adPositionRepository{db: db}
}

func (r *adPositionRepository) Create(ctx context.Context, position *model.AdPosition) error {
	if err := r.db.WithContext(ctx).Create(position).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPositionRepository) GetByID(ctx context.Context, id uint) (*model.AdPosition, error) {
	var position model.AdPosition
	if err := r.db.WithContext(ctx).First(&position, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &position, nil
}

func (r *adPositionRepository) Update(ctx context.Context, position *model.AdPosition) error {
	if err := r.db.WithContext(ctx).Model(&model.AdPosition{}).
		Where("id = ?", position.ID).
		Updates(position).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPositionRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.AdPosition{}, id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPositionRepository) List(ctx context.Context) ([]model.AdPosition, error) {
	var positions []model.AdPosition
	if err := r.db.WithContext(ctx).Order("id desc").Find(&positions).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return positions, nil
}
