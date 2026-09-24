package userservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/upload-service/config"
	"github.com/JIeeiroSst/upload-service/internal/domain/model"
	"github.com/JIeeiroSst/upload-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct {
	base string
	http *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{base: cfg.UserService.BaseURL, http: &http.Client{Timeout: cfg.UserService.Timeout}}
}

type validateResponse struct {
	Valid       bool   `json:"valid"`
	UserID      string `json:"user_id"`
	UserIDCamel string `json:"userId"`
}

func (c *Client) Validate(ctx context.Context, token string) (string, error) {
	if c.base == "" {
		return "", fmt.Errorf("%w: user-service is not configured", model.ErrUpstream)
	}
	raw, err := json.Marshal(map[string]string{"session_token": token})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/v1/validate", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: user-service: %v", model.ErrUpstream, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return "", model.ErrUnauthenticated
	case resp.StatusCode != http.StatusOK:
		return "", fmt.Errorf("%w: user-service returned status %d", model.ErrUpstream, resp.StatusCode)
	}
	var out validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("%w: user-service: %v", model.ErrUpstream, err)
	}
	id := out.UserID
	if id == "" {
		id = out.UserIDCamel
	}
	if !out.Valid || id == "" {
		return "", model.ErrUnauthenticated
	}
	return id, nil
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.TokenValidator)))),
)
