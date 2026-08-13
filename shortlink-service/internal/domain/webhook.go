package domain

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type WebhookEvent string

const (
	WebhookEventClick      WebhookEvent = "click_event"
	WebhookEventInstall    WebhookEvent = "install_event"
	WebhookEventConversion WebhookEvent = "conversion_event"
	WebhookEventSDK        WebhookEvent = "sdk_event"
)

type Webhook struct {
	ID         string
	UserID     *string
	Name       string
	URL        string
	Secret     string
	Events     []WebhookEvent
	IsActive   bool
	RetryCount int
	TimeoutMs  int
	Headers    map[string]string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type WebhookPayload struct {
	Event     WebhookEvent `json:"event"`
	EventID   string       `json:"event_id"`
	Timestamp string       `json:"timestamp"`
	Data      interface{}  `json:"data"`
}

type WebhookDeliveryResult struct {
	Success        bool
	WebhookID      string
	EventType      WebhookEvent
	EventID        string
	ResponseStatus int
	ResponseBody   string
	AttemptNumber  int
	DeliveredAt    *time.Time
	ErrorMessage   string
}

func GenerateWebhookSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func GenerateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
