package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const recipientInsertChunk = 5000

type campaignRepository struct {
	db *gorm.DB
}

func NewCampaignRepository(db *gorm.DB) port.CampaignRepository {
	return &campaignRepository{db: db}
}

func (r *campaignRepository) Create(ctx context.Context, c *model.Campaign) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *campaignRepository) AddRecipients(ctx context.Context, recipients []model.CampaignRecipient) (int64, error) {
	var inserted int64
	for start := 0; start < len(recipients); start += recipientInsertChunk {
		chunk := recipients[start:min(start+recipientInsertChunk, len(recipients))]
		res := r.db.WithContext(ctx).Clauses(clause.Insert{Modifier: "IGNORE"}).CreateInBatches(chunk, recipientInsertChunk)
		if res.Error != nil {
			return inserted, res.Error
		}
		inserted += res.RowsAffected
	}
	return inserted, nil
}

func (r *campaignRepository) Activate(ctx context.Context, id uint, recipients int) error {
	return r.db.WithContext(ctx).Model(&model.Campaign{}).Where("id = ? AND status = ?", id, model.CampaignDraft).
		Updates(map[string]any{"status": model.CampaignPending, "recipients": recipients}).Error
}

func (r *campaignRepository) GetByID(ctx context.Context, id uint) (*model.Campaign, error) {
	var c model.Campaign
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *campaignRepository) List(ctx context.Context, limit, offset int) ([]model.Campaign, error) {
	var cs []model.Campaign
	err := r.db.WithContext(ctx).Order("id DESC").Limit(limit).Offset(offset).Find(&cs).Error
	return cs, err
}

func (r *campaignRepository) SetStatus(ctx context.Context, id uint, from []model.CampaignStatus, to model.CampaignStatus, fields map[string]any) (bool, error) {
	updates := map[string]any{"status": to}
	for k, v := range fields {
		updates[k] = v
	}
	res := r.db.WithContext(ctx).Model(&model.Campaign{}).Where("id = ? AND status IN ?", id, from).Updates(updates)
	return res.RowsAffected == 1, res.Error
}

func (r *campaignRepository) Cancel(ctx context.Context, id uint) (bool, error) {
	res := r.db.WithContext(ctx).Model(&model.Campaign{}).
		Where("id = ? AND status IN ?", id, []model.CampaignStatus{model.CampaignDraft, model.CampaignPending, model.CampaignPlanning, model.CampaignSending}).
		Updates(map[string]any{"status": model.CampaignCancelled, "completed_at": time.Now().UTC()})
	return res.RowsAffected == 1, res.Error
}

func (r *campaignRepository) ClaimForPlanning(ctx context.Context, now, leaseUntil time.Time) (*model.Campaign, error) {
	var claimed *model.Campaign
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var c model.Campaign
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status IN ? AND (planner_lease_until IS NULL OR planner_lease_until < ?)",
				[]model.CampaignStatus{model.CampaignPending, model.CampaignPlanning}, now).
			Order("id ASC").First(&c).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		updates := map[string]any{"status": model.CampaignPlanning, "planner_lease_until": leaseUntil}
		if c.StartedAt == nil {
			updates["started_at"] = now
			c.StartedAt = &now
		}
		if err := tx.Model(&model.Campaign{}).Where("id = ?", c.ID).Updates(updates).Error; err != nil {
			return err
		}
		c.Status = model.CampaignPlanning
		c.PlannerLeaseUntil = &leaseUntil
		claimed = &c
		return nil
	})
	return claimed, err
}

func (r *campaignRepository) ReleasePlanning(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Campaign{}).Where("id = ?", id).Update("planner_lease_until", nil).Error
}

func (r *campaignRepository) CompleteIfDrained(ctx context.Context, id uint, now time.Time) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE notification_campaign SET status = ?, completed_at = ?, updated_at = ?
		WHERE id = ? AND status = ? AND NOT EXISTS (
			SELECT 1 FROM notification_campaign_batch b WHERE b.campaign_id = ? AND b.state IN (?, ?)
		)`, model.CampaignCompleted, now, now, id, model.CampaignSending, id, model.BatchQueued, model.BatchSending).Error
}

func (r *campaignRepository) Progress(ctx context.Context, id uint) (model.CampaignProgress, error) {
	var rows []struct {
		State   model.BatchState
		Batches int
		Targets int
		Sent    int
		Failed  int
		Invalid int
	}
	err := r.db.WithContext(ctx).Model(&model.CampaignBatch{}).
		Select("state, COUNT(*) AS batches, COALESCE(SUM(targets),0) AS targets, COALESCE(SUM(sent),0) AS sent, COALESCE(SUM(failed),0) AS failed, COALESCE(SUM(invalid),0) AS invalid").
		Where("campaign_id = ?", id).Group("state").Scan(&rows).Error
	var p model.CampaignProgress
	for _, row := range rows {
		switch row.State {
		case model.BatchQueued:
			p.BatchesQueued = row.Batches
		case model.BatchSending:
			p.BatchesSending = row.Batches
		case model.BatchSent:
			p.BatchesSent = row.Batches
		case model.BatchFailed:
			p.BatchesFailed = row.Batches
		case model.BatchCancelled:
			p.BatchesCancelled = row.Batches
		}
		p.Targets += row.Targets
		p.Sent += row.Sent
		p.Failed += row.Failed
		p.InvalidTokens += row.Invalid
	}
	return p, err
}
