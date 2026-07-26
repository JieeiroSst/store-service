package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

type notificationService struct {
	repo       port.NotificationRepository
	deviceRepo port.UserDeviceRepository
	publisher  port.NotificationPublisher
	push       port.PushSender
	email      port.EmailSender
	slack      port.SlackSender
}

func NewNotificationService(
	repo port.NotificationRepository,
	deviceRepo port.UserDeviceRepository,
	publisher port.NotificationPublisher,
	push port.PushSender,
	email port.EmailSender,
	slack port.SlackSender,
) port.NotificationUsecase {
	return &notificationService{
		repo:       repo,
		deviceRepo: deviceRepo,
		publisher:  publisher,
		push:       push,
		email:      email,
		slack:      slack,
	}
}

func (s *notificationService) CreateNotification(ctx context.Context, notification *model.Notification) (*model.Notification, error) {
	notification.Status = "pending"
	notification.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}
	if err := s.publisher.Publish(ctx, notification); err != nil {
		return nil, err
	}
	return notification, nil
}

func (s *notificationService) GetNotification(ctx context.Context, id uint) (*model.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *notificationService) ListNotifications(ctx context.Context) ([]model.Notification, error) {
	return s.repo.List(ctx)
}

func (s *notificationService) UpdateNotification(ctx context.Context, notification *model.Notification) (*model.Notification, error) {
	if _, err := s.repo.GetByID(ctx, notification.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, notification); err != nil {
		return nil, err
	}
	return notification, nil
}

func (s *notificationService) DeleteNotification(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// Dispatch delivers the notification through the channel selected by its
// Type, then records the outcome (sent/failed + retry count) on the
// notification record.
func (s *notificationService) Dispatch(ctx context.Context, notification *model.Notification) error {
	sendErr := s.send(ctx, notification)

	if sendErr != nil {
		notification.RetryCount++
		notification.Status = "failed"
		_ = s.repo.Update(ctx, notification)
		return sendErr
	}

	notification.Status = "sent"
	notification.SentAt = time.Now()
	return s.repo.Update(ctx, notification)
}

func (s *notificationService) send(ctx context.Context, notification *model.Notification) error {
	switch notification.Type {
	case "push":
		return s.sendPush(ctx, notification)
	case "email":
		return s.email.Send(ctx, []string{notification.Recipient}, notification.Title, notification.Message)
	case "slack":
		return s.slack.Send(ctx, notification.Title, notification.Message)
	default:
		return common.ErrInvalidRequest
	}
}

func (s *notificationService) sendPush(ctx context.Context, notification *model.Notification) error {
	devices, err := s.deviceRepo.ListActiveByUserID(ctx, notification.UserID)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return common.ErrNotFound
	}

	var lastErr error
	for _, device := range devices {
		if _, err := s.push.SendToToken(ctx, device.DeviceToken, notification.Title, notification.Message, nil); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
