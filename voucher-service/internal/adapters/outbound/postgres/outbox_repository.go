package postgres

import (
	"context"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"gorm.io/gorm"
)

type outboxEventModel struct {
	ID            string    `gorm:"column:id;primaryKey"`
	AggregateType string    `gorm:"column:aggregate_type"`
	AggregateID   string    `gorm:"column:aggregate_id"`
	EventType     string    `gorm:"column:event_type"`
	Payload       []byte    `gorm:"column:payload"`
	Topic         string    `gorm:"column:topic"`
	Published     bool      `gorm:"column:published"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	Attempts      int       `gorm:"column:attempts"`
	LastError     *string   `gorm:"column:last_error"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (outboxEventModel) TableName() string { return "outbox_events" }

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Enqueue(ctx context.Context, event outbox.Event) error {
	model := outboxEventModel{
		ID:            newUUID(),
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		EventType:     event.EventType,
		Payload:       event.Payload,
		Topic:         event.Topic,
		Published:     false,
		CreatedAt:     event.OccurredAt,
	}
	return txmanager.DBFromContext(ctx, r.db).Create(&model).Error
}

func (r *OutboxRepository) FetchUnpublished(ctx context.Context, limit int) ([]outbox.StoredEvent, error) {
	var models []outboxEventModel
	err := r.db.WithContext(ctx).
		Where("published = false").
		Order("created_at").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]outbox.StoredEvent, 0, len(models))
	for _, m := range models {
		out = append(out, outbox.StoredEvent{
			ID:          m.ID,
			AggregateID: m.AggregateID,
			EventType:   m.EventType,
			Payload:     m.Payload,
			Topic:       m.Topic,
			Attempts:    m.Attempts,
		})
	}
	return out, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&outboxEventModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"published": true, "published_at": now}).Error
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id, lastErr string) error {
	return r.db.WithContext(ctx).Model(&outboxEventModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"attempts": gorm.Expr("attempts + 1"), "last_error": lastErr}).Error
}
