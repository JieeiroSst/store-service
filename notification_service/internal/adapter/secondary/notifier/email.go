package notifier

import (
	"context"
	"log"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/email"
)

type emailSender struct {
	client *email.Client
}

func NewEmailSender(cfg *config.Config) port.EmailSender {
	if cfg.Email.Host == "" {
		log.Println("smtp host not configured, email notifications disabled")
		return &emailSender{}
	}
	return &emailSender{
		client: email.NewClient(cfg.Email.Host, cfg.Email.Port, cfg.Email.Username, cfg.Email.Password, cfg.Email.From),
	}
}

func (s *emailSender) Send(ctx context.Context, to []string, subject, body string) error {
	if s.client == nil {
		return common.ErrNotConfigured
	}
	return s.client.Send(to, subject, body)
}
