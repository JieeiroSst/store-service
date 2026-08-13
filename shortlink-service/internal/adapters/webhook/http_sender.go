package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
)

type HTTPSender struct{}

func NewHTTPSender() *HTTPSender { return &HTTPSender{} }

const maxResponseBodyBytes = 1000

func (s *HTTPSender) Send(ctx context.Context, wh *domain.Webhook, payload domain.WebhookPayload) domain.WebhookDeliveryResult {
	body, err := json.Marshal(payload)
	if err != nil {
		return domain.WebhookDeliveryResult{
			Success: false, WebhookID: wh.ID, EventType: payload.Event, EventID: payload.EventID,
			ErrorMessage: err.Error(),
		}
	}

	signature := domain.GenerateWebhookSignature(body, wh.Secret)

	timeout := time.Duration(wh.TimeoutMs) * time.Millisecond
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return domain.WebhookDeliveryResult{
			Success: false, WebhookID: wh.ID, EventType: payload.Event, EventID: payload.EventID,
			ErrorMessage: err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LinkForty-Signature", "sha256="+signature)
	req.Header.Set("X-LinkForty-Event", string(payload.Event))
	req.Header.Set("X-LinkForty-Event-ID", payload.EventID)
	req.Header.Set("User-Agent", "LinkForty-Webhook/1.0")
	for k, v := range wh.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		msg := err.Error()
		if reqCtx.Err() == context.DeadlineExceeded {
			msg = "Timeout after " + strconv.Itoa(wh.TimeoutMs) + "ms"
		}
		return domain.WebhookDeliveryResult{
			Success: false, WebhookID: wh.ID, EventType: payload.Event, EventID: payload.EventID,
			ErrorMessage: msg,
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	result := domain.WebhookDeliveryResult{
		Success: success, WebhookID: wh.ID, EventType: payload.Event, EventID: payload.EventID,
		ResponseStatus: resp.StatusCode, ResponseBody: string(respBody),
	}
	if success {
		now := time.Now().UTC()
		result.DeliveredAt = &now
	} else {
		result.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	return result
}
