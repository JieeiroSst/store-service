package userclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/ekyc-service/config"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type httpUserClient struct {
	client  *http.Client
	baseURL string
	token   string
}

func NewUserClient(cfg *config.Config) port.UserClient {
	return newClient(cfg)
}

func NewSessionValidator(cfg *config.Config) port.SessionValidator {
	return newClient(cfg)
}

func newClient(cfg *config.Config) *httpUserClient {
	return &httpUserClient{
		client:  &http.Client{Timeout: cfg.UserService.TimeoutDuration()},
		baseURL: strings.TrimRight(cfg.UserService.BaseURL, "/"),
		token:   cfg.UserService.Token,
	}
}

type userResponse struct {
	ID       json.Number `json:"id"`
	Username string      `json:"username"`
	Active   bool        `json:"active"`
}

func (c *httpUserClient) GetUser(ctx context.Context, userID string) (*port.UserInfo, error) {
	if _, err := strconv.ParseInt(userID, 10, 64); err != nil {
		return nil, port.ErrUserNotFound
	}
	url := fmt.Sprintf("%s/internal/v1/users/%s", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
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
	if body.ID.String() != userID {
		return nil, fmt.Errorf("user-service returned user %s for id %s", body.ID, userID)
	}
	return &port.UserInfo{ID: body.ID.String(), Username: body.Username}, nil
}

type validateResponse struct {
	Valid       bool   `json:"valid"`
	UserID      string `json:"userId"`
	UserIDSnake string `json:"user_id"`
}

func (c *httpUserClient) ValidateSession(ctx context.Context, sessionToken string) (string, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return "", port.ErrUnauthenticated
	}
	payload, _ := json.Marshal(map[string]string{"session_token": sessionToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/validate", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call user-service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return "", port.ErrUnauthenticated
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("user-service validate returned status %d", resp.StatusCode)
	}
	var v validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", fmt.Errorf("decode user-service validate response: %w", err)
	}
	id := v.UserID
	if id == "" {
		id = v.UserIDSnake
	}
	if !v.Valid || id == "" {
		return "", port.ErrUnauthenticated
	}
	return id, nil
}
