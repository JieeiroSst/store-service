package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) port.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, e *model.AuditEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *auditRepository) ListByProject(ctx context.Context, projectID uuid.UUID, entityType string, since, until *time.Time) ([]model.AuditEvent, error) {
	q := r.db.WithContext(ctx).Where("project_id = ?", projectID)
	if entityType != "" {
		q = q.Where("entity_type = ?", entityType)
	}
	if since != nil {
		q = q.Where("created_at >= ?", *since)
	}
	if until != nil {
		q = q.Where("created_at <= ?", *until)
	}

	var events []model.AuditEvent
	if err := q.Order("created_at desc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}
