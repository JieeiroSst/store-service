package userservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/auth-service/config"
	"github.com/JIeeiroSst/auth-service/internal/domain"
	"github.com/JIeeiroSst/auth-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) *Client {
		return NewClient(cfg.UserService.BaseURL, cfg.UserService.Token, cfg.UserService.Timeout)
	}),
	fx.Provide(func(c *Client) port.SessionValidator { return c }),
	fx.Provide(func(c *Client) port.UserDirectory { return c }),
)

const maxResponseBytes = 64 << 10

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: timeout}}
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Active   bool   `json:"active"`
}

func (c *Client) Username(ctx context.Context, userID int64) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/v1/users/"+strconv.FormatInt(userID, 10), nil)
	if err != nil {
		return "", err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: user-service: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("%w: user-service: %v", domain.ErrUnavailable, err)
	}
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return "", domain.Conflict("cardholder %d does not exist in user-service", userID)
	case resp.StatusCode != http.StatusOK:
		return "", fmt.Errorf("%w: user-service returned %s", domain.ErrUnavailable, resp.Status)
	}
	var u userResponse
	if err := json.Unmarshal(raw, &u); err != nil || u.ID != userID || u.Username == "" {
		return "", fmt.Errorf("%w: user-service: bad user response", domain.ErrUnavailable)
	}
	if !u.Active {
		return "", domain.Conflict("cardholder account is locked in user-service")
	}
	return u.Username, nil
}

type validateResponse struct {
	Valid       bool   `json:"valid"`
	UserID      string `json:"userId"`
	UserIDSnake string `json:"user_id"`
}

func (c *Client) Validate(ctx context.Context, sessionToken string) (int64, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return 0, domain.ErrUnauthorized
	}
	body, _ := json.Marshal(map[string]string{"session_token": sessionToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/validate", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: user-service: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return 0, fmt.Errorf("%w: user-service: %v", domain.ErrUnavailable, err)
	}
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return 0, domain.ErrUnauthorized
	case resp.StatusCode != http.StatusOK:
		return 0, fmt.Errorf("%w: user-service returned %s", domain.ErrUnavailable, resp.Status)
	}
	var v validateResponse
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0, fmt.Errorf("%w: user-service: decode: %v", domain.ErrUnavailable, err)
	}
	id := v.UserID
	if id == "" {
		id = v.UserIDSnake
	}
	userID, err := strconv.ParseInt(id, 10, 64)
	if !v.Valid || err != nil || userID <= 0 {
		return 0, domain.ErrUnauthorized
	}
	return userID, nil
}
