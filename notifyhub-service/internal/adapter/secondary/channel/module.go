package channel

import (
	"context"

	"github.com/JIeeiroSst/notifyhub-service/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/channel/email"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/channel/firebase"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/channel/sms"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(newRegistry),
)

func newRegistry(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) port.SenderRegistry {
	mail := email.New(cfg.Email, log)
	lc.Append(fx.StopHook(mail.Close))
	senders := []port.Sender{mail}
	log.Info("email channel registered", zap.String("provider", cfg.Email.Provider))

	if cfg.SMS.TwilioSID != "" || cfg.SMS.VonageKey != "" {
		senders = append(senders, sms.New(cfg.SMS, log))
		log.Info("sms channel registered", zap.String("provider", cfg.SMS.Provider))
	}

	if cfg.Firebase.CredentialsFile != "" {
		fb, err := firebase.New(context.Background(), cfg.Firebase, log)
		if err != nil {
			log.Warn("firebase init failed, channel disabled", zap.Error(err))
		} else {
			senders = append(senders, fb)
			log.Info("firebase channel registered", zap.String("project", cfg.Firebase.ProjectID))
		}
	}

	return NewRegistry(senders...)
}
