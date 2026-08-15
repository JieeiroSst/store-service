package postgres

import (
	"context"
	"encoding/json"
	"time"

	notificationapp "github.com/JIeeiroSst/voucher-service/internal/application/notification"
	"gorm.io/gorm"
)

type notificationModel struct {
	ID            string    `gorm:"column:id;primaryKey"`
	RecipientType string    `gorm:"column:recipient_type"`
	RecipientID   string    `gorm:"column:recipient_id"`
	Channel       string    `gorm:"column:channel"`
	TemplateCode  string    `gorm:"column:template_code"`
	Payload       []byte    `gorm:"column:payload"`
	Status        string    `gorm:"column:status"`
	Error         *string   `gorm:"column:error"`
	SentAt        *time.Time `gorm:"column:sent_at"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (notificationModel) TableName() string { return "notifications" }

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) notificationapp.NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, id, recipientType, recipientID string, channel notificationapp.Channel, templateCode string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	model := notificationModel{
		ID:            id,
		RecipientType: recipientType,
		RecipientID:   recipientID,
		Channel:       string(channel),
		TemplateCode:  templateCode,
		Payload:       body,
		Status:        "queued",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *NotificationRepository) MarkSent(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&notificationModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": "sent", "sent_at": now, "updated_at": now}).Error
}

func (r *NotificationRepository) MarkFailed(ctx context.Context, id, errMsg string) error {
	return r.db.WithContext(ctx).Model(&notificationModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": "failed", "error": errMsg, "updated_at": time.Now().UTC()}).Error
}
