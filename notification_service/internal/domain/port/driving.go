package port

import (
	"context"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
)

type NotificationUsecase interface {
	CreateNotification(ctx context.Context, notification *model.Notification) (*model.Notification, error)
	GetNotification(ctx context.Context, id uint) (*model.Notification, error)
	ListNotifications(ctx context.Context) ([]model.Notification, error)
	UpdateNotification(ctx context.Context, notification *model.Notification) (*model.Notification, error)
	DeleteNotification(ctx context.Context, id uint) error

	Dispatch(ctx context.Context, notification *model.Notification) error
}

type UserDeviceUsecase interface {
	RegisterDevice(ctx context.Context, device *model.UserDevice) (*model.UserDevice, error)
	GetDevice(ctx context.Context, id uint) (*model.UserDevice, error)
	ListDevices(ctx context.Context) ([]model.UserDevice, error)
	UpdateDevice(ctx context.Context, device *model.UserDevice) (*model.UserDevice, error)
	DeleteDevice(ctx context.Context, id uint) error
}
