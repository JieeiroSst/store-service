package notifierclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/doordash-service/config"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type httpNotifierClient struct {
	client  *http.Client
	baseURL string
}

func NewNotifierClient(cfg *config.Config) port.NotifierClient {
	return &httpNotifierClient{
		client:  &http.Client{Timeout: cfg.NotificationService.TimeoutDuration()},
		baseURL: cfg.NotificationService.BaseURL,
	}
}

type notifyRequest struct {
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	Type        string `json:"type"`
	ReferenceID string `json:"reference_id"`
}

func (c *httpNotifierClient) Notify(ctx context.Context, in port.NotifyInput) error {
	payload, err := json.Marshal(notifyRequest{
		UserID: in.UserID, Title: in.Title, Message: in.Message, Type: in.Type, ReferenceID: in.ReferenceID,
	})
	if err != nil {
		return fmt.Errorf("encode notification request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/notifications", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call notification service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}
	return nil
}
