package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

type notificationService struct {
	repo          port.NotificationRepository
	deviceRepo    port.UserDeviceRepository
	contacts      port.ContactRepository
	deactivator   port.DeviceDeactivator
	publisher     port.NotificationPublisher
	push          port.MulticastSender
	email         port.EmailSender
	slack         port.SlackSender
	template      port.TemplateRenderer
	slackTemplate port.SlackTemplateRenderer
	audit         *auditor
}

func NewNotificationService(
	repo port.NotificationRepository,
	deviceRepo port.UserDeviceRepository,
	contacts port.ContactRepository,
	deactivator port.DeviceDeactivator,
	publisher port.NotificationPublisher,
	push port.MulticastSender,
	email port.EmailSender,
	slack port.SlackSender,
	template port.TemplateRenderer,
	slackTemplate port.SlackTemplateRenderer,
	audit *auditor,
) port.NotificationUsecase {
	return &notificationService{
		repo:          repo,
		deviceRepo:    deviceRepo,
		contacts:      contacts,
		deactivator:   deactivator,
		publisher:     publisher,
		push:          push,
		email:         email,
		slack:         slack,
		template:      template,
		slackTemplate: slackTemplate,
		audit:         audit,
	}
}

func (s *notificationService) CreateNotification(ctx context.Context, notification *model.Notification) (*model.Notification, error) {
	notification.RequestedBy = strings.TrimSpace(notification.RequestedBy)
	if notification.RequestedBy == "" {
		notification.RequestedBy = model.DefaultRequester
	}
	if model.RawDataSize(notification.RawData) > model.MaxRawDataBytes {
		return nil, invalid("raw_data must be at most %d bytes of JSON", model.MaxRawDataBytes)
	}
	switch notification.Type {
	case "email":
		if strings.TrimSpace(notification.Recipient) == "" {
			return nil, invalid("email needs a recipient")
		}
		if notification.TemplateType == "" && strings.TrimSpace(notification.Title) == "" {
			return nil, invalid("email needs template_type, or a subject for a raw email")
		}
		if notification.TemplateType == "" && strings.TrimSpace(notification.Message) == "" && len(notification.RawData) == 0 {
			return nil, invalid("raw email needs html, text or raw_data")
		}
	case "slack":
		if notification.TemplateType == "" && strings.TrimSpace(notification.Title) == "" && strings.TrimSpace(notification.Message) == "" && len(notification.RawData) == 0 {
			return nil, invalid("slack needs template_type, or title, text or raw_data")
		}
	}
	if notification.Type == "push" {
		notification.Email = model.NormalizeEmail(notification.Email)
		if strings.TrimSpace(notification.Phone) != "" {
			phone, ok := model.NormalizePhone(notification.Phone)
			if !ok {
				return nil, invalid("phone %q is not valid", notification.Phone)
			}
			notification.Phone = phone
		}
		if notification.UserID == 0 && notification.Email == "" && notification.Phone == "" {
			return nil, invalid("push needs user_id, email or phone")
		}
		if strings.TrimSpace(notification.Title) == "" && strings.TrimSpace(notification.Message) == "" {
			return nil, invalid("push needs a title or message")
		}
	}
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

func (s *notificationService) Dispatch(ctx context.Context, notification *model.Notification, attempt int, final bool) error {
	t := newTrail(model.AuditSourceNotification, notification.ID, 0, attempt, notification.Type, notification.RequestedBy)
	t.setContent(notification.Title, notification.Message, notification.Data, notification.TemplateType)
	sendErr := s.send(ctx, notification, t)
	if len(t.deliveries) == 0 {
		t.add(model.AuditDelivery{UserID: notification.UserID, Recipient: auditRecipient(notification)})
	}
	if sendErr != nil {
		t.failAll(sendErr)
	}
	for i := range t.deliveries {
		if t.deliveries[i].Status == "" {
			t.deliveries[i].Status = model.AuditSent
		}
	}
	s.audit.record(ctx, t)
	notification.RetryCount = max(attempt-1, 0)

	if sendErr == nil {
		notification.Status = "sent"
		sentAt := time.Now()
		notification.SentAt = &sentAt
		notification.LastError = ""
		return s.repo.Update(ctx, notification)
	}

	permanent := IsPermanent(sendErr)
	notification.LastError = sendErr.Error()
	notification.Status = "retrying"
	if permanent || final {
		notification.Status = "failed"
	}
	if err := s.repo.Update(ctx, notification); err != nil {
		return err
	}
	if permanent {
		return fmt.Errorf("%w: %v", common.ErrPermanent, sendErr)
	}
	return sendErr
}

func IsPermanent(err error) bool {
	return errors.Is(err, common.ErrPermanent) ||
		errors.Is(err, common.ErrNoActiveDevice) ||
		errors.Is(err, common.ErrInvalidRequest) ||
		errors.Is(err, common.ErrNotConfigured) ||
		errors.Is(err, common.ErrNotFound)
}

func auditRecipient(n *model.Notification) string {
	switch {
	case n.Recipient != "":
		return n.Recipient
	case n.Email != "":
		return n.Email
	default:
		return n.Phone
	}
}

func (s *notificationService) send(ctx context.Context, notification *model.Notification, t *trail) error {
	switch notification.Type {
	case "push":
		return s.sendPush(ctx, notification, t)
	case "email":
		return s.sendEmail(ctx, notification, t)
	case "slack":
		return s.sendSlack(ctx, notification, t)
	default:
		return common.ErrInvalidRequest
	}
}

// sendSlack sends Title/Message as-is unless TemplateType is set, in which
// case the title and mrkdwn text are rendered from that template with
// TemplateData instead.
func (s *notificationService) sendSlack(ctx context.Context, notification *model.Notification, t *trail) error {
	title, text := notification.Title, notification.Message

	if notification.TemplateType != "" {
		renderedTitle, renderedText, err := s.slackTemplate.Render(notification.TemplateType, notification.TemplateData)
		if err != nil {
			return fmt.Errorf("%w: %v", common.ErrPermanent, err)
		}
		title, text = renderedTitle, renderedText
	}
	if block := model.RawDataSlack(notification.RawData); block != "" {
		text = strings.TrimSpace(text + "\n" + block)
	}
	t.setContent(title, text, nil, notification.TemplateType)
	t.add(model.AuditDelivery{UserID: notification.UserID, Recipient: "slack"})
	return s.slack.Send(ctx, title, text)
}

// sendEmail sends Title/Message as-is unless TemplateType is set, in which
// case the subject and HTML body are rendered from that template with
// TemplateData instead.
func (s *notificationService) sendEmail(ctx context.Context, notification *model.Notification, t *trail) error {
	subject, body := notification.Title, notification.Message

	if notification.TemplateType != "" {
		renderedSubject, renderedHTML, err := s.template.Render(notification.TemplateType, notification.TemplateData)
		if err != nil {
			return fmt.Errorf("%w: %v", common.ErrPermanent, err)
		}
		subject, body = renderedSubject, renderedHTML
	}
	body += model.RawDataHTML(notification.RawData)
	t.setContent(subject, body, nil, notification.TemplateType)
	t.add(model.AuditDelivery{UserID: notification.UserID, Recipient: model.NormalizeEmail(notification.Recipient)})
	return s.email.Send(ctx, []string{notification.Recipient}, subject, body)
}

func (s *notificationService) resolveUserID(ctx context.Context, n *model.Notification) (uint, error) {
	switch {
	case n.UserID != 0:
		return n.UserID, nil
	case n.Email != "":
		return s.contacts.UserIDByEmail(ctx, n.Email)
	case n.Phone != "":
		return s.contacts.UserIDByPhone(ctx, n.Phone)
	default:
		return 0, fmt.Errorf("%w: push needs user_id, email or phone", common.ErrInvalidRequest)
	}
}

func (s *notificationService) sendPush(ctx context.Context, notification *model.Notification, t *trail) error {
	userID, err := s.resolveUserID(ctx, notification)
	if err != nil {
		if errors.Is(err, common.ErrUnknownContact) {
			return fmt.Errorf("%w: %v", common.ErrPermanent, err)
		}
		return err
	}
	devices, err := s.deviceRepo.ListActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		t.add(model.AuditDelivery{UserID: userID, Recipient: auditRecipient(notification)})
		return common.ErrNoActiveDevice
	}

	tokens := make([]string, len(devices))
	for i, d := range devices {
		tokens[i] = d.DeviceToken
	}
	results, err := s.push.SendToTokens(ctx, tokens, notification.Title, notification.Message, notification.Data)
	for i, d := range devices {
		entry := model.AuditDelivery{UserID: userID, DeviceID: d.ID, DeviceType: d.DeviceType, TokenHash: model.HashToken(d.DeviceToken)}
		if err == nil && i < len(results) {
			entry.Status, entry.MessageID, entry.Error = pushStatus(results[i]), results[i].MessageID, results[i].Error
		}
		t.add(entry)
	}
	if err != nil {
		return err
	}

	var (
		unregistered []uint
		sent         int
		retryable    bool
		lastErr      string
	)
	for i, r := range results {
		switch {
		case r.Success:
			sent++
		case r.Unregistered:
			unregistered = append(unregistered, devices[i].ID)
		default:
			retryable = retryable || r.Retryable
			lastErr = r.Error
		}
	}
	if err := s.deactivator.DeactivateByIDs(ctx, unregistered); err != nil {
		return err
	}

	switch {
	case sent > 0:
		return nil
	case len(unregistered) == len(devices):
		return common.ErrNoActiveDevice
	case retryable:
		return errors.New(lastErr)
	default:
		return fmt.Errorf("%w: %s", common.ErrPermanent, lastErr)
	}
}

func pushStatus(r port.TokenResult) string {
	switch {
	case r.Success:
		return model.AuditSent
	case r.Unregistered:
		return model.AuditInvalidToken
	default:
		return model.AuditFailed
	}
}
