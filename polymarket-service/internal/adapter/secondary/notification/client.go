package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/polymarket-service/config"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
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

func (c *Client) Notify(ctx context.Context, n port.Notification) error {
	if c.baseURL == "" {
		return nil
	}
	userID, _ := strconv.ParseUint(n.UserID, 10, 64)
	raw, err := json.Marshal(map[string]any{
		"user_id":   userID,
		"recipient": n.UserID,
		"title":     n.Title,
		"message":   n.Message,
		"type":      "push",
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/notifications", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notification-service returned status %d", resp.StatusCode)
	}
	return nil
}
