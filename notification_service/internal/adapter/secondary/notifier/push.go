package notifier

import (
	"context"
	"fmt"
	"log"

	"firebase.google.com/go/v4/messaging"
	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/firebase"
)

type pushSender struct {
	client *firebase.FirebaseMessaging
}

func NewPushSender(cfg *config.Config) (*pushSender, error) {
	if cfg.Firebase.CredentialsFile == "" {
		log.Println("firebase credentials not configured, push notifications disabled")
		return &pushSender{}, nil
	}

	client, err := firebase.NewFirebaseMessaging(cfg.Firebase.CredentialsFile)
	if err != nil {
		return nil, err
	}
	return &pushSender{client: client}, nil
}

func (s *pushSender) SendToToken(ctx context.Context, token, title, body string, data map[string]string) (string, error) {
	if s.client == nil {
		return "", common.ErrNotConfigured
	}
	return s.client.SendToToken(ctx, token, title, body, data)
}

func (s *pushSender) SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) ([]port.TokenResult, error) {
	if s.client == nil {
		return nil, common.ErrNotConfigured
	}
	resp, err := s.client.SendEachForMulticast(ctx, tokens, title, body, data)
	if err != nil {
		return nil, err
	}
	results := make([]port.TokenResult, len(tokens))
	for i, r := range resp.Responses {
		if i >= len(results) {
			break
		}
		results[i] = classify(r.Success, r.Error)
		results[i].MessageID = r.MessageID
	}
	return results, nil
}

func classify(success bool, err error) port.TokenResult {
	if success || err == nil {
		return port.TokenResult{Success: true}
	}
	return port.TokenResult{
		Unregistered: messaging.IsUnregistered(err),
		Retryable: messaging.IsUnavailable(err) || messaging.IsInternal(err) || messaging.IsQuotaExceeded(err) ||
			messaging.IsMessageRateExceeded(err) || messaging.IsServerUnavailable(err) || messaging.IsUnknown(err),
		Error: err.Error(),
	}
}

func (s *pushSender) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) (string, error) {
	if s.client == nil {
		return "", common.ErrNotConfigured
	}
	return s.client.SendToTopic(ctx, topic, title, body, data)
}

func (s *pushSender) ValidateToken(ctx context.Context, token string) error {
	if s.client == nil {
		return common.ErrNotConfigured
	}
	_, err := s.client.ValidateToken(ctx, token)
	if err != nil && (messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) || messaging.IsSenderIDMismatch(err)) {
		return fmt.Errorf("%w: %v", common.ErrInvalidToken, err)
	}
	return err
}
