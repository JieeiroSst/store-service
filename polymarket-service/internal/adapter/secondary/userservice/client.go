package userservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/polymarket-service/config"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.UserService.BaseURL, "/"),
		http:    &http.Client{Timeout: cfg.UserService.TimeoutDuration()},
	}
}

type validateResponse struct {
	Valid       bool   `json:"valid"`
	UserID      string `json:"user_id"`
	UserIDCamel string `json:"userId"`
}

func (c *Client) Validate(ctx context.Context, token string) (string, error) {
	if c.baseURL == "" {
		return "", fmt.Errorf("%w: user-service is not configured", port.ErrUpstream)
	}
	raw, err := json.Marshal(map[string]string{"session_token": token})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/validate", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", port.ErrUpstream, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return "", port.ErrUnauthenticated
	case resp.StatusCode != http.StatusOK:
		return "", fmt.Errorf("%w: user-service returned status %d", port.ErrUpstream, resp.StatusCode)
	}
	var out validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("%w: %v", port.ErrUpstream, err)
	}
	id := out.UserID
	if id == "" {
		id = out.UserIDCamel
	}
	if !out.Valid || id == "" {
		return "", port.ErrUnauthenticated
	}
	return id, nil
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.UserDirectory)))),
)
