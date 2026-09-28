package notifier

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/email"
)

type emailSender struct {
	client *email.Client
}

func NewEmailSender(cfg *config.Config) *emailSender {
	if cfg.Email.APIKey == "" {
		log.Println("resend api key not configured, email notifications disabled")
		return &emailSender{}
	}
	return &emailSender{
		client: email.NewClient(cfg.Email.APIKey, cfg.Email.From),
	}
}

func (s *emailSender) Send(ctx context.Context, to []string, subject, body string) error {
	if s.client == nil {
		return common.ErrNotConfigured
	}
	err := s.client.Send(to, subject, body)
	var apiErr *email.APIError
	if errors.As(err, &apiErr) && !apiErr.Retryable() {
		return fmt.Errorf("%w: %v", common.ErrPermanent, err)
	}
	return err
}

func (s *emailSender) SendBatch(ctx context.Context, messages []port.EmailMessage, idempotencyKey string) error {
	if s.client == nil {
		return common.ErrNotConfigured
	}
	batch := make([]email.Message, len(messages))
	for i, m := range messages {
		batch[i] = email.Message{To: m.To, Subject: m.Subject, HTML: m.HTML}
	}
	err := s.client.SendBatch(ctx, batch, idempotencyKey)
	var apiErr *email.APIError
	if errors.As(err, &apiErr) && !apiErr.Retryable() {
		return fmt.Errorf("%w: %v", common.ErrPermanent, err)
	}
	return err
}

func (s *emailSender) MaxBatchSize() int { return email.MaxBatchSize }
