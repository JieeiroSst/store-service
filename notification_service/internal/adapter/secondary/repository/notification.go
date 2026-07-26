package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) port.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uint) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.WithContext(ctx).First(&notification, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &notification, nil
}

func (r *notificationRepository) Update(ctx context.Context, notification *model.Notification) error {
	if err := r.db.WithContext(ctx).Model(&model.Notification{}).Where("id = ?", notification.ID).Updates(notification).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Notification{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *notificationRepository) List(ctx context.Context) ([]model.Notification, error) {
	var notifications []model.Notification
	if err := r.db.WithContext(ctx).Find(&notifications).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return notifications, nil
}
