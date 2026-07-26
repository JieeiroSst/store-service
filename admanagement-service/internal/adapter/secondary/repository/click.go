package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
	"gorm.io/gorm"
)

type adClickRepository struct {
	db *gorm.DB
}

func NewAdClickRepository(db *gorm.DB) port.AdClickRepository {
	return &adClickRepository{db: db}
}

func (r *adClickRepository) Create(ctx context.Context, click *model.AdClick) error {
	if err := r.db.WithContext(ctx).Create(click).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adClickRepository) GetByID(ctx context.Context, id uint) (*model.AdClick, error) {
	var click model.AdClick
	if err := r.db.WithContext(ctx).First(&click, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &click, nil
}

func (r *adClickRepository) List(ctx context.Context) ([]model.AdClick, error) {
	var clicks []model.AdClick
	if err := r.db.WithContext(ctx).Order("id desc").Find(&clicks).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return clicks, nil
}

func (r *adClickRepository) ListByAd(ctx context.Context, adID uint) ([]model.AdClick, error) {
	var clicks []model.AdClick
	if err := r.db.WithContext(ctx).Where("ad_id = ?", adID).Order("id desc").Find(&clicks).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return clicks, nil
}

func (r *adClickRepository) CountByAdAndDate(ctx context.Context, adID uint, date time.Time) (int64, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	var count int64
	if err := r.db.WithContext(ctx).Model(&model.AdClick{}).
		Where("ad_id = ? AND created_at >= ? AND created_at < ?", adID, start, end).
		Count(&count).Error; err != nil {
		return 0, common.ErrDBFailed
	}
	return count, nil
}
