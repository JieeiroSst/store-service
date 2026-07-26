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

type adImpressionRepository struct {
	db *gorm.DB
}

func NewAdImpressionRepository(db *gorm.DB) port.AdImpressionRepository {
	return &adImpressionRepository{db: db}
}

func (r *adImpressionRepository) Create(ctx context.Context, impression *model.AdImpression) error {
	if err := r.db.WithContext(ctx).Create(impression).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adImpressionRepository) GetByID(ctx context.Context, id uint) (*model.AdImpression, error) {
	var impression model.AdImpression
	if err := r.db.WithContext(ctx).First(&impression, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &impression, nil
}

func (r *adImpressionRepository) List(ctx context.Context) ([]model.AdImpression, error) {
	var impressions []model.AdImpression
	if err := r.db.WithContext(ctx).Order("id desc").Find(&impressions).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return impressions, nil
}

func (r *adImpressionRepository) ListByAd(ctx context.Context, adID uint) ([]model.AdImpression, error) {
	var impressions []model.AdImpression
	if err := r.db.WithContext(ctx).Where("ad_id = ?", adID).Order("id desc").Find(&impressions).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return impressions, nil
}

func (r *adImpressionRepository) CountByAdAndDate(ctx context.Context, adID uint, date time.Time) (int64, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	var count int64
	if err := r.db.WithContext(ctx).Model(&model.AdImpression{}).
		Where("ad_id = ? AND created_at >= ? AND created_at < ?", adID, start, end).
		Count(&count).Error; err != nil {
		return 0, common.ErrDBFailed
	}
	return count, nil
}
