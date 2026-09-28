package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"gorm.io/gorm"
)

type historyRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) port.HistoryRepository {
	return &historyRepository{db: db}
}

func (r *historyRepository) Create(ctx context.Context, h *model.NotifyHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *historyRepository) UpdateStatus(ctx context.Context, id string, status model.NotifyStatus, retries int, errMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      status,
		"retry_count": retries,
		"error":       errMsg,
		"updated_at":  now,
	}
	if status == model.NotifyStatusSent {
		updates["sent_at"] = now
	}
	return r.db.WithContext(ctx).
		Model(&model.NotifyHistory{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *historyRepository) List(ctx context.Context, f port.HistoryFilter) ([]*model.NotifyHistory, int64, error) {
	filter := func(q *gorm.DB) *gorm.DB {
		if f.JobID != "" {
			q = q.Where("job_id = ?", f.JobID)
		}
		if f.Status != "" {
			q = q.Where("status = ?", f.Status)
		}
		return q
	}

	var total int64
	if err := filter(r.db.WithContext(ctx).Model(&model.NotifyHistory{})).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var hist []*model.NotifyHistory
	err := filter(r.db.WithContext(ctx)).
		Order("created_at desc").
		Offset(offset(f.Page, f.PageSize)).
		Limit(f.PageSize).
		Find(&hist).Error
	return hist, total, err
}
