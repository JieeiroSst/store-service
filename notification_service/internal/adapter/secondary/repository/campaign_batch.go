package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"gorm.io/gorm"
)

type campaignBatchRepository struct {
	db *gorm.DB
}

func NewCampaignBatchRepository(db *gorm.DB) port.CampaignBatchRepository {
	return &campaignBatchRepository{db: db}
}

func (r *campaignBatchRepository) CreateAndAdvance(ctx context.Context, b *model.CampaignBatch, cursor uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(b).Error; err != nil {
			return err
		}
		return tx.Model(&model.Campaign{}).Where("id = ?", b.CampaignID).Updates(map[string]any{
			"cursor_id":       cursor,
			"batches_planned": gorm.Expr("batches_planned + 1"),
		}).Error
	})
}

func (r *campaignBatchRepository) GetByID(ctx context.Context, id uint) (*model.CampaignBatch, error) {
	var b model.CampaignBatch
	if err := r.db.WithContext(ctx).First(&b, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *campaignBatchRepository) Claim(ctx context.Context, id uint, now, leaseUntil time.Time) (*model.CampaignBatch, bool, error) {
	res := r.db.WithContext(ctx).Model(&model.CampaignBatch{}).
		Where("id = ? AND (state = ? OR (state = ? AND lease_until < ?))", id, model.BatchQueued, model.BatchSending, now).
		Updates(map[string]any{"state": model.BatchSending, "lease_until": leaseUntil, "attempts": gorm.Expr("attempts + 1")})
	if res.Error != nil || res.RowsAffected == 0 {
		return nil, false, res.Error
	}
	b, err := r.GetByID(ctx, id)
	return b, err == nil, err
}

func (r *campaignBatchRepository) Finish(ctx context.Context, b *model.CampaignBatch) error {
	return r.db.WithContext(ctx).Model(&model.CampaignBatch{}).Where("id = ?", b.ID).Updates(map[string]any{
		"state":       b.State,
		"sent":        b.Sent,
		"failed":      b.Failed,
		"invalid":     b.Invalid,
		"last_error":  b.LastError,
		"lease_until": nil,
	}).Error
}

func (r *campaignBatchRepository) Requeue(ctx context.Context, id uint, republishAfter time.Time, lastErr string) error {
	return r.db.WithContext(ctx).Model(&model.CampaignBatch{}).Where("id = ?", id).Updates(map[string]any{
		"state":           model.BatchQueued,
		"lease_until":     nil,
		"republish_after": republishAfter,
		"last_error":      lastErr,
	}).Error
}

func (r *campaignBatchRepository) DueForRepublish(ctx context.Context, now time.Time, limit int) ([]model.CampaignBatch, error) {
	var bs []model.CampaignBatch
	err := r.db.WithContext(ctx).
		Where("(state = ? AND republish_after < ?) OR (state = ? AND lease_until < ?)", model.BatchQueued, now, model.BatchSending, now).
		Order("id ASC").Limit(limit).Find(&bs).Error
	return bs, err
}

func (r *campaignBatchRepository) TouchRepublish(ctx context.Context, id uint, republishAfter time.Time) error {
	return r.db.WithContext(ctx).Model(&model.CampaignBatch{}).Where("id = ?", id).Update("republish_after", republishAfter).Error
}

func (r *campaignBatchRepository) CancelPending(ctx context.Context, campaignID uint) error {
	return r.db.WithContext(ctx).Model(&model.CampaignBatch{}).
		Where("campaign_id = ? AND state = ?", campaignID, model.BatchQueued).
		Update("state", model.BatchCancelled).Error
}
