package userclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/ekyc-service/config"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type httpUserClient struct {
	client  *http.Client
	baseURL string
}

func NewUserClient(cfg *config.Config) port.UserClient {
	return &httpUserClient{
		client:  &http.Client{Timeout: cfg.UserService.TimeoutDuration()},
		baseURL: cfg.UserService.BaseURL,
	}
}

type userResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (c *httpUserClient) GetUser(ctx context.Context, userID string) (*port.UserInfo, error) {
	url := fmt.Sprintf("%s/v1/users/%s", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call user-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, port.ErrUserNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service returned status %d for user %s", resp.StatusCode, userID)
	}

	var body userResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode user-service response: %w", err)
	}
	return &port.UserInfo{ID: body.ID, Username: body.Username}, nil
}
