package repository

import (
	"context"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type webhookEventRepository struct {
	db *gorm.DB
}

func NewWebhookEventRepository(db *gorm.DB) port.WebhookEventRepository {
	return &webhookEventRepository{db: db}
}

func (r *webhookEventRepository) Exists(ctx context.Context, webhookID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.WebhookEvent{}).Where("webhook_id = ?", webhookID).Count(&count).Error
	return count > 0, err
}

func (r *webhookEventRepository) Create(ctx context.Context, event *model.WebhookEvent) error {
	// A concurrent retry of the same delivery may have recorded it first.
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(event).Error
}
