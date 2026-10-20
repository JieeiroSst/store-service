package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FraudRepository struct {
	db *gorm.DB
}

func NewFraudRepository(db *gorm.DB) *FraudRepository {
	return &FraudRepository{db: db}
}

func toBlock(rec blockRecord) model.UserBlock {
	return model.UserBlock{UserID: rec.UserID, Reason: rec.Reason, BlockedBy: rec.BlockedBy, CreatedAt: rec.CreatedAt.UTC()}
}

func (r *FraudRepository) BlockOf(ctx context.Context, userID string) (*model.UserBlock, error) {
	var rec blockRecord
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, translate(err)
	}
	b := toBlock(rec)
	return &b, nil
}

func (r *FraudRepository) Block(ctx context.Context, b *model.UserBlock) (int64, error) {
	var held int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rec := blockRecord{UserID: b.UserID, Reason: b.Reason, BlockedBy: b.BlockedBy, CreatedAt: time.Now().UTC()}
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"reason", "blocked_by"}),
		}).Create(&rec).Error
		if err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", b.UserID).Take(&rec).Error; err != nil {
			return err
		}
		*b = toBlock(rec)
		res := tx.Model(&outboxRecord{}).
			Where("user_id = ? AND status = ? AND published_at IS NULL", b.UserID, string(model.RewardPending)).
			Updates(map[string]any{"status": string(model.RewardHeld), "hold_reason": model.HoldBlockedUser})
		held = res.RowsAffected
		return res.Error
	})
	return held, translate(err)
}

func (r *FraudRepository) Unblock(ctx context.Context, userID string) error {
	res := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&blockRecord{})
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *FraudRepository) Blocks(ctx context.Context, limit, offset int) ([]model.UserBlock, int64, error) {
	q := r.db.WithContext(ctx).Model(&blockRecord{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	var recs []blockRecord
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&recs).Error; err != nil {
		return nil, 0, translate(err)
	}
	out := make([]model.UserBlock, 0, len(recs))
	for _, rec := range recs {
		out = append(out, toBlock(rec))
	}
	return out, total, nil
}

func (r *FraudRepository) Rewards(ctx context.Context, f port.RewardFilter) ([]model.ReviewReward, int64, error) {
	q := r.db.WithContext(ctx).Model(&outboxRecord{}).Where("status = ?", string(f.Status))
	if f.Status == model.RewardPending {
		q = q.Where("published_at IS NULL")
	}
	if f.UserID != "" {
		q = q.Where("user_id = ?", f.UserID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	var recs []outboxRecord
	if err := q.Order("id DESC").Limit(f.Limit).Offset(f.Offset).Find(&recs).Error; err != nil {
		return nil, 0, translate(err)
	}
	out := make([]model.ReviewReward, 0, len(recs))
	for _, rec := range recs {
		out = append(out, toReviewReward(rec))
	}
	return out, total, nil
}

func (r *FraudRepository) Review(ctx context.Context, id int64, from []model.RewardStatus, to model.RewardStatus, by string) (*model.ReviewReward, error) {
	statuses := make([]string, 0, len(from))
	for _, s := range from {
		statuses = append(statuses, string(s))
	}
	var rec outboxRecord
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&rec).Error; err != nil {
			return err
		}
		if rec.PublishedAt != nil || !contains(statuses, rec.Status) {
			return model.ErrConflict
		}
		now := time.Now().UTC()
		rec.Status, rec.ReviewedBy, rec.ReviewedAt = string(to), by, &now
		err := tx.Model(&outboxRecord{}).Where("id = ?", id).Updates(map[string]any{
			"status":      rec.Status,
			"reviewed_by": by,
			"reviewed_at": now,
		}).Error
		if err != nil || to != model.RewardRejected || rec.Source == "lucky-checkin" {
			return err
		}
		return tx.Exec("UPDATE lucky_campaigns SET points_awarded = GREATEST(points_awarded - ?, 0) WHERE id = ?", rec.Points, rec.CampaignID).Error
	})
	if err != nil {
		return nil, translate(err)
	}
	out := toReviewReward(rec)
	return &out, nil
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
