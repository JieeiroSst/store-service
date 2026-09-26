package notificationservice

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/ticket-service/config"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/httpx"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type Client struct{ c *httpx.Client }

var (
	_ outbound.PushGateway  = (*Client)(nil)
	_ outbound.EmailGateway = (*Client)(nil)
)

func NewClient(cfg config.Config) *Client {
	return &Client{c: httpx.New("notification-service", cfg.NotificationServiceURL, cfg.PaymentServiceTimeout)}
}

func (c *Client) Push(ctx context.Context, n domain.Notification) error {
	st, msg, err := c.c.Do(ctx, http.MethodPost, "/api/v1/notifications", nil, map[string]any{
		"user_id": n.UserID, "title": n.Title, "message": n.Body, "type": "push", "priority": 1,
	}, nil)
	if err != nil {
		return err
	}
	if !httpx.OK(st) {
		return fmt.Errorf("%w: notification-service returned %d: %s", domain.ErrUpstreamUnavailable, st, msg)
	}
	return nil
}

func (c *Client) Email(ctx context.Context, n domain.Notification) error {
	if n.Email == nil {
		return nil
	}
	st, msg, err := c.c.Do(ctx, http.MethodPost, "/api/v1/notifications/email", nil, map[string]any{
		"user_id": n.UserID, "recipient": n.Email.To, "template_type": n.Email.Template, "template_data": n.Email.Data, "priority": 1,
	}, nil)
	if err != nil {
		return err
	}
	if !httpx.OK(st) {
		return fmt.Errorf("%w: notification-service returned %d: %s", domain.ErrUpstreamUnavailable, st, msg)
	}
	return nil
}
