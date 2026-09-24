package notification

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/httpx"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct{ http *httpx.Client }

func New(cfg *config.Config) *Client {
	return &Client{http: httpx.New(cfg.Notification.BaseURL, cfg.Notification.Timeout)}
}

// Notify is a no-op when notification-service is not configured.
func (c *Client) Notify(ctx context.Context, n port.Notification) error {
	if !c.http.Enabled() {
		return nil
	}
	kind := "push"
	if n.Email != "" {
		kind = "email"
	}
	err := c.http.Do(ctx, http.MethodPost, "/api/v1/notifications", map[string]any{
		"user_id":   n.PatientID,
		"recipient": recipient(n),
		"title":     n.Title,
		"message":   n.Message,
		"type":      kind,
	}, nil)
	if err != nil {
		return fmt.Errorf("%w: notification-service: %v", model.ErrUpstream, err)
	}
	return nil
}

func recipient(n port.Notification) string {
	if n.Email != "" {
		return n.Email
	}
	return fmt.Sprint(n.PatientID)
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.Notifier)))),
)
