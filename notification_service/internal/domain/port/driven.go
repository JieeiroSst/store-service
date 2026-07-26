package port

import (
	"context"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	GetByID(ctx context.Context, id uint) (*model.Notification, error)
	Update(ctx context.Context, notification *model.Notification) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.Notification, error)
}

type UserDeviceRepository interface {
	Create(ctx context.Context, device *model.UserDevice) error
	GetByID(ctx context.Context, id uint) (*model.UserDevice, error)
	ListActiveByUserID(ctx context.Context, userID uint) ([]model.UserDevice, error)
	Update(ctx context.Context, device *model.UserDevice) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.UserDevice, error)
}

type NotificationPublisher interface {
	Publish(ctx context.Context, notification *model.Notification) error
}

type PushSender interface {
	SendToToken(ctx context.Context, token, title, body string, data map[string]string) (string, error)
}

type EmailSender interface {
	Send(ctx context.Context, to []string, subject, body string) error
}

type SlackSender interface {
	Send(ctx context.Context, title, message string) error
}
