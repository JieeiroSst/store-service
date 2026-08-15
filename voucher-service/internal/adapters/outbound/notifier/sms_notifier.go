package notifier

import (
	"context"

	notificationapp "github.com/JIeeiroSst/voucher-service/internal/application/notification"
	"go.uber.org/zap"
)

type SMSNotifier struct {
	log *zap.Logger
}

func NewSMSNotifier(log *zap.Logger) notificationapp.Notifier {
	return &SMSNotifier{log: log}
}

func (n *SMSNotifier) Channel() notificationapp.Channel { return notificationapp.ChannelSMS }

func (n *SMSNotifier) Send(ctx context.Context, recipient, templateCode string, payload map[string]any) error {
	n.log.Info("sms notification (stand-in, no SMS provider configured)",
		zap.String("recipient", recipient),
		zap.String("template", templateCode),
		zap.Any("payload", payload),
	)
	return nil
}
