package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
	"gorm.io/gorm"
)

type adPositionMappingRepository struct {
	db *gorm.DB
}

func NewAdPositionMappingRepository(db *gorm.DB) port.AdPositionMappingRepository {
	return &adPositionMappingRepository{db: db}
}

func (r *adPositionMappingRepository) Create(ctx context.Context, mapping *model.AdPositionMapping) error {
	if err := r.db.WithContext(ctx).Create(mapping).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPositionMappingRepository) GetByID(ctx context.Context, id uint) (*model.AdPositionMapping, error) {
	var mapping model.AdPositionMapping
	if err := r.db.WithContext(ctx).First(&mapping, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &mapping, nil
}

func (r *adPositionMappingRepository) Update(ctx context.Context, mapping *model.AdPositionMapping) error {
	if err := r.db.WithContext(ctx).Model(&model.AdPositionMapping{}).
		Where("id = ?", mapping.ID).
		Updates(mapping).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPositionMappingRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.AdPositionMapping{}, id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPositionMappingRepository) List(ctx context.Context) ([]model.AdPositionMapping, error) {
	var mappings []model.AdPositionMapping
	if err := r.db.WithContext(ctx).Order("id desc").Find(&mappings).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return mappings, nil
}

func (r *adPositionMappingRepository) ListActiveByPosition(ctx context.Context, positionID uint) ([]model.AdPositionMapping, error) {
	var mappings []model.AdPositionMapping
	if err := r.db.WithContext(ctx).
		Where("position_id = ? AND is_active = ?", positionID, true).
		Find(&mappings).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return mappings, nil
}
