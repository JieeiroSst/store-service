package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

type notificationRequest struct {
	Recipient string `json:"recipient,omitempty"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Priority  int    `json:"priority"`
}

type httpNotifier struct {
	client    *http.Client
	endpoint  string
	kind      string
	recipient string
}

func New(cfg *config.Config) port.Notifier {
	n := cfg.Notification
	if n.URL == "" {
		return noop{}
	}
	return &httpNotifier{
		client:    &http.Client{Timeout: n.Timeout},
		endpoint:  n.URL + "/api/v1/notifications",
		kind:      n.Type,
		recipient: n.Recipient,
	}
}

func (h *httpNotifier) Notify(ctx context.Context, msg model.Notification) error {
	recipient := msg.Recipient
	if recipient == "" {
		recipient = h.recipient
	}
	body, err := json.Marshal(notificationRequest{
		Recipient: recipient,
		Title:     msg.Title,
		Message:   msg.Message,
		Type:      h.kind,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("notification_service returned %s", resp.Status)
	}
	return nil
}

type noop struct{}

func (noop) Notify(context.Context, model.Notification) error { return nil }
