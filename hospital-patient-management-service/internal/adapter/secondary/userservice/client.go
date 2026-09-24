package userservice

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
	return &Client{http: httpx.New(cfg.UserService.BaseURL, cfg.UserService.Timeout)}
}

type validateResponse struct {
	Valid       bool   `json:"valid"`
	UserID      string `json:"user_id"`
	UserIDCamel string `json:"userId"`
}

func (c *Client) Validate(ctx context.Context, token string) (string, error) {
	if !c.http.Enabled() {
		return "", fmt.Errorf("%w: user-service is not configured", model.ErrUpstream)
	}
	var out validateResponse
	err := c.http.Do(ctx, http.MethodPost, "/api/v1/validate", map[string]string{"session_token": token}, &out)
	switch {
	case httpx.IsStatus(err, http.StatusUnauthorized), httpx.IsStatus(err, http.StatusForbidden):
		return "", model.ErrUnauthenticated
	case err != nil:
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
