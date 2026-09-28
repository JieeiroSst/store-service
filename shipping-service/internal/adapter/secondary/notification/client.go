package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.Notification.BaseURL, "/"),
		http:    &http.Client{Timeout: cfg.Notification.TimeoutDuration()},
	}
}

func (c *Client) Notify(ctx context.Context, n model.Notification) error {
	if c.baseURL == "" {
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"user_id":   n.UserID,
		"recipient": n.Recipient,
		"title":     n.Title,
		"message":   n.Message,
		"type":      n.Channel,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/notifications", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("notification-service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("notification-service: status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.Notifier)))),
)
