package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type adPerformanceSummaryRepository struct {
	db *gorm.DB
}

func NewAdPerformanceSummaryRepository(db *gorm.DB) port.AdPerformanceSummaryRepository {
	return &adPerformanceSummaryRepository{db: db}
}

func (r *adPerformanceSummaryRepository) Create(ctx context.Context, summary *model.AdPerformanceSummary) error {
	if err := r.db.WithContext(ctx).Create(summary).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPerformanceSummaryRepository) GetByID(ctx context.Context, id uint) (*model.AdPerformanceSummary, error) {
	var summary model.AdPerformanceSummary
	if err := r.db.WithContext(ctx).First(&summary, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &summary, nil
}

func (r *adPerformanceSummaryRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.AdPerformanceSummary{}, id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPerformanceSummaryRepository) List(ctx context.Context) ([]model.AdPerformanceSummary, error) {
	var summaries []model.AdPerformanceSummary
	if err := r.db.WithContext(ctx).Order("date desc").Find(&summaries).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return summaries, nil
}

func (r *adPerformanceSummaryRepository) GetByAdAndDate(ctx context.Context, adID uint, date time.Time) (*model.AdPerformanceSummary, error) {
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	var summary model.AdPerformanceSummary
	if err := r.db.WithContext(ctx).
		Where("ad_id = ? AND date = ?", adID, day).
		First(&summary).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &summary, nil
}

func (r *adPerformanceSummaryRepository) Upsert(ctx context.Context, summary *model.AdPerformanceSummary) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "ad_id"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"impressions", "clicks", "ctr", "cost", "revenue", "updated_at"}),
		}).
		Create(summary).Error
	if err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adPerformanceSummaryRepository) ListByCampaign(ctx context.Context, campaignID uint, from, to time.Time) ([]model.AdPerformanceSummary, error) {
	var summaries []model.AdPerformanceSummary
	if err := r.db.WithContext(ctx).
		Joins("JOIN ads ON ads.id = ad_performance_summary.ad_id").
		Where("ads.campaign_id = ? AND ad_performance_summary.date >= ? AND ad_performance_summary.date <= ?", campaignID, from, to).
		Find(&summaries).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return summaries, nil
}
