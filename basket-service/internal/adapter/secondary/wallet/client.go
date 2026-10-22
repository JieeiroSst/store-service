package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/config"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct {
	baseURL    string
	merchantID string
	http       *http.Client
}

func New(baseURL, merchantID string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{baseURL: baseURL, merchantID: merchantID, http: &http.Client{Timeout: timeout}}
}

var Module = fx.Options(fx.Provide(func(cfg *config.Config) port.Wallet {
	w := cfg.WalletService
	return New(w.BaseURL, w.MerchantID, time.Duration(w.TimeoutSeconds)*time.Second)
}))

func (c *Client) Pay(ctx context.Context, p port.WalletPayment) (string, error) {
	payload, err := json.Marshal(map[string]any{
		"user_id":         strconv.Itoa(p.UserID),
		"wallet_id":       p.WalletID,
		"merchant_id":     c.merchantID,
		"amount":          strconv.FormatFloat(p.Amount, 'f', 2, 64),
		"pin":             p.Pin,
		"description":     p.Description,
		"metadata":        p.Metadata,
		"idempotency_key": p.IdempotencyKey,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/payments", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", p.IdempotencyKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: wallet: %v", common.ErrUpstream, err)
	}
	defer resp.Body.Close()
	var body struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Error  string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	switch {
	case resp.StatusCode >= 500:
		return "", fmt.Errorf("%w: wallet returned %d", common.ErrUpstream, resp.StatusCode)
	case resp.StatusCode >= 400:
		msg := body.Error
		if msg == "" {
			msg = fmt.Sprintf("status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("%w: %s", common.ErrPaymentFailed, msg)
	case body.ID == "":
		return "", fmt.Errorf("%w: wallet returned no transaction id", common.ErrUpstream)
	}
	return body.ID, nil
}
