package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const auditInsertChunk = 500

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) port.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) SaveContent(ctx context.Context, c *model.AuditContent) error {
	db := r.db.WithContext(ctx)
	if err := db.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(c).Error; err != nil {
		return err
	}
	var stored model.AuditContent
	if err := db.Select("id").Where("hash = ?", c.Hash).First(&stored).Error; err != nil {
		return err
	}
	c.ID = stored.ID
	return nil
}

func (r *auditRepository) SaveDeliveries(ctx context.Context, deliveries []model.AuditDelivery) error {
	if len(deliveries) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(deliveries, auditInsertChunk).Error
}

func (r *auditRepository) ListDeliveries(ctx context.Context, f port.AuditFilter) ([]model.AuditDelivery, error) {
	q := r.db.WithContext(ctx).Model(&model.AuditDelivery{})
	switch {
	case len(f.UserIDs) > 0 && len(f.Recipients) > 0:
		q = q.Where("(user_id IN ? OR recipient IN ?)", f.UserIDs, f.Recipients)
	case len(f.UserIDs) > 0:
		q = q.Where("user_id IN ?", f.UserIDs)
	case len(f.Recipients) > 0:
		q = q.Where("recipient IN ?", f.Recipients)
	}
	if f.Channel != "" {
		q = q.Where("channel = ?", f.Channel)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.SourceType != "" {
		q = q.Where("source_type = ?", f.SourceType)
	}
	if f.SourceID != 0 {
		q = q.Where("source_id = ?", f.SourceID)
	}
	if f.RequestedBy != "" {
		q = q.Where("requested_by = ?", f.RequestedBy)
	}
	if !f.From.IsZero() {
		q = q.Where("created_at >= ?", f.From)
	}
	if !f.To.IsZero() {
		q = q.Where("created_at < ?", f.To)
	}
	if f.BeforeID != 0 {
		q = q.Where("id < ?", f.BeforeID)
	}
	var out []model.AuditDelivery
	err := q.Order("id DESC").Limit(f.Limit).Find(&out).Error
	return out, err
}

func (r *auditRepository) GetContents(ctx context.Context, ids []uint) ([]model.AuditContent, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var out []model.AuditContent
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&out).Error
	return out, err
}

func (r *auditRepository) DeleteBefore(ctx context.Context, before time.Time, limit int) (int64, error) {
	res := r.db.WithContext(ctx).Exec("DELETE FROM notification_audit_delivery WHERE created_at < ? ORDER BY id LIMIT ?", before, limit)
	return res.RowsAffected, res.Error
}
