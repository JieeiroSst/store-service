package app

import (
	"context"
	"math"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"go.uber.org/zap"
)

type WebhookTrigger struct {
	sender ports.WebhookSender
	log    *zap.Logger
}

func NewWebhookTrigger(sender ports.WebhookSender, log *zap.Logger) *WebhookTrigger {
	return &WebhookTrigger{sender: sender, log: log}
}

func (t *WebhookTrigger) Trigger(ctx context.Context, webhooks []*domain.Webhook, event domain.WebhookEvent, eventID string, data interface{}) {
	payload := domain.WebhookPayload{
		Event:     event,
		EventID:   eventID,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Data:      data,
	}

	relevant := make([]*domain.Webhook, 0, len(webhooks))
	for _, w := range webhooks {
		if !w.IsActive {
			continue
		}
		for _, e := range w.Events {
			if e == event {
				relevant = append(relevant, w)
				break
			}
		}
	}
	if len(relevant) == 0 {
		return
	}

	bgCtx := context.Background()
	for _, w := range relevant {
		go t.deliverWithRetry(bgCtx, w, payload)
	}
}

func (t *WebhookTrigger) deliverWithRetry(ctx context.Context, w *domain.Webhook, payload domain.WebhookPayload) domain.WebhookDeliveryResult {
	maxRetries := w.RetryCount
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var result domain.WebhookDeliveryResult
	for attempt := 1; attempt <= maxRetries; attempt++ {
		result = t.sender.Send(ctx, w, payload)
		result.AttemptNumber = attempt

		if result.Success {
			return result
		}

		if attempt < maxRetries {
			delay := time.Duration(math.Min(1000*math.Pow(2, float64(attempt-1)), 30000)) * time.Millisecond
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return result
			}
		}
	}

	t.log.Error("webhook delivery failed after retries",
		zap.String("webhookId", w.ID), zap.String("event", string(payload.Event)),
		zap.String("error", result.ErrorMessage))
	return result
}
