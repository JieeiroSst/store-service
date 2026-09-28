package ghn

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type Client struct {
	baseURL      string
	token        string
	webhookToken string
	http         *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:      strings.TrimRight(cfg.GHN.BaseURL, "/"),
		token:        cfg.GHN.Token,
		webhookToken: cfg.GHN.WebhookToken,
		http:         &http.Client{Timeout: cfg.GHN.TimeoutDuration()},
	}
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) do(ctx context.Context, shopID int64, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Token", c.token)
	if shopID > 0 {
		req.Header.Set("ShopId", strconv.FormatInt(shopID, 10))
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: ghn %s: %v", port.ErrCarrierUnavailable, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))

	var env envelope
	_ = json.Unmarshal(raw, &env)
	msg := env.Message
	if msg == "" {
		msg = strings.TrimSpace(string(raw[:min(len(raw), 300)]))
	}

	switch {
	case resp.StatusCode >= 500:
		return fmt.Errorf("%w: ghn %s: status %d: %s", port.ErrCarrierUnavailable, path, resp.StatusCode, msg)
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("%w: ghn %s: authentication failed: %s", port.ErrCarrierUnavailable, path, msg)
	case resp.StatusCode >= 400 || (env.Code != 0 && env.Code != http.StatusOK):
		return fmt.Errorf("%w: ghn: %s", port.ErrCarrierRejected, msg)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("%w: ghn %s: decode: %v", port.ErrCarrierUnavailable, path, err)
	}
	return nil
}

func (c *Client) VerifyWebhookToken(token string) bool {
	if c.webhookToken == "" || token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(c.webhookToken)) == 1
}

func parseTime(raw string) *time.Time {
	if raw == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	t = t.UTC()
	return &t
}
