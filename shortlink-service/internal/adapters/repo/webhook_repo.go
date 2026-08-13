package repo

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type WebhookRepo struct{ db *gorm.DB }

func NewWebhookRepo(db *gorm.DB) *WebhookRepo { return &WebhookRepo{db: db} }

func webhookToModel(w *domain.Webhook) *WebhookModel {
	events := make(pq.StringArray, len(w.Events))
	for i, e := range w.Events {
		events[i] = string(e)
	}
	headers := map[string]interface{}{}
	for k, v := range w.Headers {
		headers[k] = v
	}
	m := &WebhookModel{
		ID: w.ID, UserID: w.UserID, Name: w.Name, URL: w.URL, Secret: w.Secret,
		Events: events, IsActive: w.IsActive, RetryCount: w.RetryCount, TimeoutMs: w.TimeoutMs,
		Headers: mapToJSON(headers),
	}
	return m
}

func modelToWebhook(m *WebhookModel) *domain.Webhook {
	events := make([]domain.WebhookEvent, len(m.Events))
	for i, e := range m.Events {
		events[i] = domain.WebhookEvent(e)
	}
	rawHeaders := jsonToMap(m.Headers)
	headers := map[string]string{}
	for k, v := range rawHeaders {
		if s, ok := v.(string); ok {
			headers[k] = s
		}
	}
	return &domain.Webhook{
		ID: m.ID, UserID: m.UserID, Name: m.Name, URL: m.URL, Secret: m.Secret,
		Events: events, IsActive: m.IsActive, RetryCount: m.RetryCount, TimeoutMs: m.TimeoutMs,
		Headers: headers, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *WebhookRepo) Create(ctx context.Context, webhook *domain.Webhook) error {
	m := webhookToModel(webhook)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*webhook = *modelToWebhook(m)
	return nil
}

func (r *WebhookRepo) scoped(ctx context.Context, userID *string) *gorm.DB {
	q := r.db.WithContext(ctx)
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	return q
}

func (r *WebhookRepo) List(ctx context.Context, userID *string) ([]*domain.Webhook, error) {
	var models []WebhookModel
	if err := r.scoped(ctx, userID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Webhook, len(models))
	for i := range models {
		out[i] = modelToWebhook(&models[i])
	}
	return out, nil
}

func (r *WebhookRepo) GetByID(ctx context.Context, id string, userID *string) (*domain.Webhook, error) {
	var m WebhookModel
	err := r.scoped(ctx, userID).Where("id = ?", id).Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return modelToWebhook(&m), nil
}

func (r *WebhookRepo) Update(ctx context.Context, id string, userID *string, patch map[string]interface{}) (*domain.Webhook, error) {
	if len(patch) == 0 {
		return nil, errors.New("no fields to update")
	}
	res := r.scoped(ctx, userID).Model(&WebhookModel{}).Where("id = ?", id).Updates(patch)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id, nil)
}

func (r *WebhookRepo) Delete(ctx context.Context, id string, userID *string) error {
	res := r.scoped(ctx, userID).Delete(&WebhookModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *WebhookRepo) ActiveForUser(ctx context.Context, userID string) ([]*domain.Webhook, error) {
	var models []WebhookModel
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = true", userID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Webhook, len(models))
	for i := range models {
		out[i] = modelToWebhook(&models[i])
	}
	return out, nil
}

func (r *WebhookRepo) ActiveForLinkOwner(ctx context.Context, linkID string) ([]*domain.Webhook, error) {
	var models []WebhookModel
	err := r.db.WithContext(ctx).Table("webhooks w").
		Select("w.*").
		Joins("INNER JOIN links l ON l.user_id = w.user_id").
		Where("l.id = ? AND w.is_active = true", linkID).
		Scan(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Webhook, len(models))
	for i := range models {
		out[i] = modelToWebhook(&models[i])
	}
	return out, nil
}
