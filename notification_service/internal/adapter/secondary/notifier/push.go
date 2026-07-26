package notifier

import (
	"context"
	"log"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/firebase"
)

type pushSender struct {
	client *firebase.FirebaseMessaging
}

func NewPushSender(cfg *config.Config) (port.PushSender, error) {
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
