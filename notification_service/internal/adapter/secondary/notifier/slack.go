package notifier

import (
	"context"
	"fmt"
	"log"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/slack"
)

type slackSender struct {
	client  slack.Slack
	channel string
}

func NewSlackSender(cfg *config.Config) port.SlackSender {
	if cfg.Slack.WebhookSecret == "" {
		log.Println("slack webhook not configured, slack notifications disabled")
		return &slackSender{}
	}
	return &slackSender{
		client:  slack.NewSlack(cfg.Slack.WebhookSecret),
		channel: cfg.Slack.Channel,
	}
}

func (s *slackSender) Send(ctx context.Context, title, message string) error {
	if s.client == nil {
		return common.ErrNotConfigured
	}
	return s.client.PushNoti(slack.PayloadSlack{
		Channel:  s.channel,
		Username: "notification-service",
		Text:     fmt.Sprintf("*%s*\n%s", title, message),
	})
}
