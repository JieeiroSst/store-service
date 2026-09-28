package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) port.OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Enqueue(ctx context.Context, jobs []model.OutboxJob) error {
	if len(jobs) == 0 {
		return nil
	}
	return conn(ctx, r.db).Create(&jobs).Error
}

func (r *outboxRepository) ClaimDue(ctx context.Context, now time.Time, limit int, lease time.Duration) ([]model.OutboxJob, error) {
	var jobs []model.OutboxJob
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("state = ? AND next_attempt_at <= ?", model.JobPending, now).
			Order("id ASC").Limit(limit).Find(&jobs).Error
		if err != nil || len(jobs) == 0 {
			return err
		}
		ids := make([]int64, len(jobs))
		for i := range jobs {
			ids[i] = jobs[i].ID
			jobs[i].Attempts++
		}
		return tx.Model(&model.OutboxJob{}).Where("id IN ?", ids).Updates(map[string]any{
			"attempts":        gorm.Expr("attempts + 1"),
			"next_attempt_at": now.Add(lease),
			"updated_at":      now,
		}).Error
	})
	return jobs, err
}

func (r *outboxRepository) MarkDone(ctx context.Context, id int64) error {
	return conn(ctx, r.db).Model(&model.OutboxJob{}).Where("id = ?", id).
		Updates(map[string]any{"state": model.JobDone, "last_error": ""}).Error
}

func (r *outboxRepository) MarkRetry(ctx context.Context, id int64, next time.Time, lastErr string) error {
	return conn(ctx, r.db).Model(&model.OutboxJob{}).Where("id = ?", id).
		Updates(map[string]any{"next_attempt_at": next, "last_error": truncate(lastErr)}).Error
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id int64, lastErr string) error {
	return conn(ctx, r.db).Model(&model.OutboxJob{}).Where("id = ?", id).
		Updates(map[string]any{"state": model.JobFailed, "last_error": truncate(lastErr)}).Error
}

func truncate(s string) string {
	if len(s) > 1000 {
		return s[:1000]
	}
	return s
}
