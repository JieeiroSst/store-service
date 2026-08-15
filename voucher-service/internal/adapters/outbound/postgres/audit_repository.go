package postgres

import (
	"context"
	"encoding/json"
	"time"

	auditapp "github.com/JIeeiroSst/voucher-service/internal/application/audit"
	"gorm.io/gorm"
)

type auditLogModel struct {
	ID         string    `gorm:"column:id;primaryKey"`
	ActorType  string    `gorm:"column:actor_type"`
	ActorID    string    `gorm:"column:actor_id"`
	Action     string    `gorm:"column:action"`
	EntityType string    `gorm:"column:entity_type"`
	EntityID   string    `gorm:"column:entity_id"`
	Before     []byte    `gorm:"column:before"`
	After      []byte    `gorm:"column:after"`
	IPAddress  string    `gorm:"column:ip_address"`
	UserAgent  string    `gorm:"column:user_agent"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (auditLogModel) TableName() string { return "audit_log" }

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) auditapp.AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Insert(ctx context.Context, entry auditapp.Entry) error {
	before, err := json.Marshal(entry.Before)
	if err != nil {
		return err
	}
	after, err := json.Marshal(entry.After)
	if err != nil {
		return err
	}
	model := auditLogModel{
		ID:         newUUID(),
		ActorType:  entry.ActorType,
		ActorID:    entry.ActorID,
		Action:     entry.Action,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		Before:     before,
		After:      after,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		CreatedAt:  time.Now().UTC(),
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *AuditRepository) Query(ctx context.Context, in auditapp.QueryInput) ([]auditapp.Entry, error) {
	q := r.db.WithContext(ctx).Model(&auditLogModel{})
	if in.EntityType != "" {
		q = q.Where("entity_type = ?", in.EntityType)
	}
	if in.EntityID != "" {
		q = q.Where("entity_id = ?", in.EntityID)
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 100
	}
	var models []auditLogModel
	if err := q.Order("created_at DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]auditapp.Entry, 0, len(models))
	for _, m := range models {
		var before, after map[string]any
		_ = json.Unmarshal(m.Before, &before)
		_ = json.Unmarshal(m.After, &after)
		out = append(out, auditapp.Entry{
			ActorType:  m.ActorType,
			ActorID:    m.ActorID,
			Action:     m.Action,
			EntityType: m.EntityType,
			EntityID:   m.EntityID,
			Before:     before,
			After:      after,
			IPAddress:  m.IPAddress,
			UserAgent:  m.UserAgent,
			CreatedAt:  m.CreatedAt,
		})
	}
	return out, nil
}
