package otp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/auth-service/config"
	"github.com/JIeeiroSst/auth-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewSender),
)

func NewSender(cfg *config.Config) (port.OTPSender, error) {
	switch cfg.OTPDelivery.Backend {
	case "log":
		log.Printf("otp delivery: log (development only, codes are written to the log)")
		return LogSender{}, nil
	case "webhook":
		return NewWebhook(cfg.OTPDelivery.WebhookURL, cfg.OTPDelivery.WebhookToken, cfg.OTPDelivery.Timeout), nil
	}
	return nil, fmt.Errorf("unknown OTP_DELIVERY %q (log or webhook)", cfg.OTPDelivery.Backend)
}

type LogSender struct{}

func (LogSender) Send(_ context.Context, m port.OTPMessage) error {
	log.Printf("[otp] user=%d challenge=%s code=%s amount=%d %s merchant=%q card=%s expires=%s",
		m.UserID, m.ChallengeID, m.OTP, m.Amount, m.Currency, m.Merchant, m.MaskedPAN, m.ExpiresAt.Format(time.RFC3339))
	return nil
}

type Webhook struct {
	url   string
	token string
	http  *http.Client
}

func NewWebhook(url, token string, timeout time.Duration) *Webhook {
	return &Webhook{url: url, token: token, http: &http.Client{Timeout: timeout}}
}

type webhookPayload struct {
	UserID      int64     `json:"user_id"`
	Channel     string    `json:"channel"`
	Template    string    `json:"template"`
	ChallengeID string    `json:"challenge_id"`
	OTP         string    `json:"otp"`
	Message     string    `json:"message"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (w *Webhook) Send(ctx context.Context, m port.OTPMessage) error {
	body, err := json.Marshal(webhookPayload{
		UserID:      m.UserID,
		Channel:     "otp",
		Template:    "payment_authentication",
		ChallengeID: m.ChallengeID,
		OTP:         m.OTP,
		Message: fmt.Sprintf("Ma OTP %s xac thuc thanh toan %s %s tai %s bang the %s. Hieu luc den %s. Khong chia se ma nay.",
			m.OTP, strconv.FormatInt(m.Amount, 10), m.Currency, m.Merchant, m.MaskedPAN, m.ExpiresAt.Format("15:04")),
		ExpiresAt: m.ExpiresAt,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if w.token != "" {
		req.Header.Set("Authorization", "Bearer "+w.token)
	}
	resp, err := w.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("otp webhook returned %s", resp.Status)
	}
	return nil
}
