// Package referral is the driven adapter onto referral-service's REST API.
package referral

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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
		baseURL: strings.TrimRight(cfg.Referral.BaseURL, "/"),
		http:    &http.Client{Timeout: cfg.Referral.TimeoutDuration()},
	}
}

// referral-service wraps every response as {"data": ...} or {"error": {...}}.
type envelope struct {
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) GenerateLink(ctx context.Context, ownerUserID string) (*port.ReferralLink, error) {
	var out struct {
		RefCode   string `json:"ref_code"`
		DeepLink  string `json:"deep_link"`
		ExpiresAt int64  `json:"expires_at"` // unix ms
	}
	err := c.do(ctx, http.MethodPost, "/api/v1/referral/generate",
		map[string]string{"owner_user_id": ownerUserID, "channel": "copy", "platform": "universal"}, &out)
	if err != nil {
		return nil, err
	}
	return &port.ReferralLink{RefCode: out.RefCode, DeepLink: out.DeepLink, ExpiresAt: time.UnixMilli(out.ExpiresAt)}, nil
}

func (c *Client) Activate(ctx context.Context, refCode, userID string) (string, error) {
	var out struct {
		Attributed  bool   `json:"attributed"`
		OwnerUserID string `json:"owner_user_id"`
	}
	err := c.do(ctx, http.MethodPost, "/api/v1/referral/activate",
		map[string]string{"ref_code": refCode, "user_id": userID}, &out)
	if err != nil {
		return "", err
	}
	if !out.Attributed {
		return "", fmt.Errorf("%w: referral code could not be applied", port.ErrInvalidInput)
	}
	return out.OwnerUserID, nil
}

func (c *Client) Stats(ctx context.Context, userID string) (*port.ReferralStats, error) {
	var out port.ReferralStats
	if err := c.do(ctx, http.MethodGet, "/api/v1/referral/user/"+url.PathEscape(userID)+"/stats", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	if c.baseURL == "" {
		return fmt.Errorf("%w: referral-service is not configured", port.ErrUpstream)
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", port.ErrUpstream, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var env envelope
	_ = json.Unmarshal(raw, &env)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return json.Unmarshal(env.Data, out)
	}
	return statusError(resp.StatusCode, env)
}

// statusError maps referral-service's error codes: a bad or unusable code is
// the caller's problem, anything else is an upstream failure.
func statusError(status int, env envelope) error {
	msg := "referral-service returned status " + fmt.Sprint(status)
	code := ""
	if env.Error != nil {
		msg, code = env.Error.Message, env.Error.Code
	}
	switch {
	case code == "SELF_REFERRAL" || code == "ALREADY_REFERRED":
		return port.ErrReferralNotAllowed
	case status == http.StatusNotFound || status == http.StatusUnprocessableEntity || status == http.StatusBadRequest:
		return fmt.Errorf("%w: %s", port.ErrInvalidInput, msg)
	default:
		return fmt.Errorf("%w: %s", port.ErrUpstream, msg)
	}
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.ReferralGateway)))),
)
