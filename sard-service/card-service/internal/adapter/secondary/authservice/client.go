package authservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JIeeiroSst/card-service/config"
	"github.com/JIeeiroSst/card-service/internal/domain"
	"github.com/JIeeiroSst/card-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) port.PaymentAuthenticator {
		if cfg.AuthService.BaseURL == "" {
			log.Printf("auth-service is not configured: ECOM step-up authentication is not enforced")
			return Disabled{}
		}
		return NewClient(cfg.AuthService.BaseURL, cfg.AuthService.Token, cfg.AuthService.Timeout)
	}),
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: timeout}}
}

func (c *Client) Enabled() bool { return true }

func (c *Client) Consume(ctx context.Context, authenticationID, cardID string, amount int64, currency string) error {
	body, _ := json.Marshal(map[string]any{"card_id": cardID, "amount": amount, "currency": currency})
	endpoint := c.baseURL + "/internal/v1/payment-authentications/" + url.PathEscape(authenticationID) + "/consume"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: auth-service: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return domain.ErrNotFound
	case http.StatusConflict, http.StatusGone, http.StatusBadRequest:
		return domain.Conflict("payment authentication rejected: %s", strings.TrimSpace(string(raw)))
	}
	return fmt.Errorf("%w: auth-service returned %s", domain.ErrUnavailable, resp.Status)
}

type Disabled struct{}

func (Disabled) Enabled() bool { return false }

func (Disabled) Consume(context.Context, string, string, int64, string) error { return nil }
