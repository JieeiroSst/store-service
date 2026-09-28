package port

import (
	"context"
	"time"

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
	Register(ctx context.Context, device *model.UserDevice, opts RegisterOptions) (*model.UserDevice, error)
	DeactivateByTokenHash(ctx context.Context, tokenHash string) (int64, error)
	DeactivateByUserID(ctx context.Context, userID uint) (int64, error)
	DeactivateStale(ctx context.Context, lastUsedBefore time.Time) (int64, error)
	DeactivateByDeviceID(ctx context.Context, userID uint, deviceID string) (int64, error)
	ListByUserID(ctx context.Context, userID uint) ([]model.UserDevice, error)
}

type ContactRepository interface {
	Upsert(ctx context.Context, contact *model.UserContact) error
	Get(ctx context.Context, userID uint) (*model.UserContact, error)
	UserIDByEmail(ctx context.Context, email string) (uint, error)
	UserIDByPhone(ctx context.Context, phone string) (uint, error)
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) error
}

type RegisterOptions struct {
	ReplaceSameInstall  bool
	SingleDevicePerUser bool
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

// TemplateRenderer renders a named email template (subject + HTML body),
// substituting placeholders from data. templateType selects the template
// file (see internal/adapter/secondary/template/templates).
type TemplateRenderer interface {
	Render(templateType string, data map[string]string) (subject string, html string, err error)
}

// SlackTemplateRenderer renders a named Slack message template (title +
// mrkdwn text), substituting placeholders from data. templateType selects
// the template file (see internal/adapter/secondary/slacktemplate/templates).
type SlackTemplateRenderer interface {
	Render(templateType string, data map[string]string) (title string, text string, err error)
}

type TokenResult struct {
	MessageID    string
	Success      bool
	Unregistered bool
	Retryable    bool
	Error        string
}

type MulticastSender interface {
	SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) ([]TokenResult, error)
	SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) (string, error)
}

type EmailMessage struct {
	To      string
	Subject string
	HTML    string
}

type BatchEmailSender interface {
	SendBatch(ctx context.Context, messages []EmailMessage, idempotencyKey string) error
	MaxBatchSize() int
}

type DeviceDeactivator interface {
	DeactivateByIDs(ctx context.Context, ids []uint) error
}

type CampaignRepository interface {
	Create(ctx context.Context, c *model.Campaign) error
	AddRecipients(ctx context.Context, recipients []model.CampaignRecipient) (int64, error)
	Activate(ctx context.Context, id uint, recipients int) error
	GetByID(ctx context.Context, id uint) (*model.Campaign, error)
	List(ctx context.Context, limit, offset int) ([]model.Campaign, error)
	SetStatus(ctx context.Context, id uint, from []model.CampaignStatus, to model.CampaignStatus, fields map[string]any) (bool, error)
	Cancel(ctx context.Context, id uint) (bool, error)
	ClaimForPlanning(ctx context.Context, now, leaseUntil time.Time) (*model.Campaign, error)
	ReleasePlanning(ctx context.Context, id uint) error
	CompleteIfDrained(ctx context.Context, id uint, now time.Time) error
	Progress(ctx context.Context, id uint) (model.CampaignProgress, error)
}

type CampaignTargetSource interface {
	NextIDs(ctx context.Context, c *model.Campaign, afterID uint, limit int) ([]uint, error)
	Targets(ctx context.Context, c *model.Campaign, afterID, untilID uint) ([]model.BatchTarget, error)
}

type CampaignBatchRepository interface {
	CreateAndAdvance(ctx context.Context, b *model.CampaignBatch, cursor uint) error
	GetByID(ctx context.Context, id uint) (*model.CampaignBatch, error)
	Claim(ctx context.Context, id uint, now, leaseUntil time.Time) (*model.CampaignBatch, bool, error)
	Finish(ctx context.Context, b *model.CampaignBatch) error
	Requeue(ctx context.Context, id uint, republishAfter time.Time, lastErr string) error
	DueForRepublish(ctx context.Context, now time.Time, limit int) ([]model.CampaignBatch, error)
	TouchRepublish(ctx context.Context, id uint, republishAfter time.Time) error
	CancelPending(ctx context.Context, campaignID uint) error
}

type BatchPublisher interface {
	PublishBatch(ctx context.Context, batchID uint) error
}

type Clock interface {
	Now() time.Time
}

type AuditFilter struct {
	UserIDs     []uint
	Recipients  []string
	Channel     string
	Status      string
	SourceType  string
	SourceID    uint
	RequestedBy string
	From        time.Time
	To          time.Time
	BeforeID    uint
	Limit       int
}

type AuditRepository interface {
	SaveContent(ctx context.Context, c *model.AuditContent) error
	SaveDeliveries(ctx context.Context, deliveries []model.AuditDelivery) error
	ListDeliveries(ctx context.Context, f AuditFilter) ([]model.AuditDelivery, error)
	GetContents(ctx context.Context, ids []uint) ([]model.AuditContent, error)
	DeleteBefore(ctx context.Context, before time.Time, limit int) (int64, error)
}
