package notification

import "context"

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
)

type SendInput struct {
	RecipientType string
	RecipientID   string
	Channel       Channel
	TemplateCode  string
	Payload       map[string]any
}

type NotificationService interface {
	Send(ctx context.Context, in SendInput) error
}

type NotificationRepository interface {
	Create(ctx context.Context, id, recipientType, recipientID string, channel Channel, templateCode string, payload map[string]any) error
	MarkSent(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id, errMsg string) error
}

type Notifier interface {
	Channel() Channel
	Send(ctx context.Context, recipient, templateCode string, payload map[string]any) error
}

type NotifierRegistry interface {
	Resolve(channel Channel) (Notifier, error)
}
