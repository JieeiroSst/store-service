package cardservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/auth-service/config"
	"github.com/JIeeiroSst/auth-service/internal/domain"
	"github.com/JIeeiroSst/auth-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) port.CardDirectory {
		return NewClient(cfg.CardService.BaseURL, cfg.CardService.Token, cfg.CardService.Timeout)
	}),
)

const maxResponseBytes = 1 << 20

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: timeout}}
}

type resolveResponse struct {
	CardID     string `json:"card_id"`
	AccountID  string `json:"account_id"`
	CustomerID string `json:"customer_id"`
	UserID     int64  `json:"user_id"`
	MaskedPAN  string `json:"masked_pan"`
	Status     string `json:"status"`
}

func (c *Client) ResolvePAN(ctx context.Context, pan, expiry string) (domain.CardRef, error) {
	body, _ := json.Marshal(map[string]string{"pan": pan, "expiry": expiry})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/cards/resolve", bytes.NewReader(body))
	if err != nil {
		return domain.CardRef{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return domain.CardRef{}, fmt.Errorf("%w: card-service: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return domain.CardRef{}, fmt.Errorf("%w: card-service: %v", domain.ErrUnavailable, err)
	}
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return domain.CardRef{}, domain.Conflict("card is not enrolled for payment authentication")
	case resp.StatusCode == http.StatusBadRequest:
		return domain.CardRef{}, domain.Invalid("card-service rejected the card data: %s", strings.TrimSpace(string(raw)))
	case resp.StatusCode != http.StatusOK:
		return domain.CardRef{}, fmt.Errorf("%w: card-service returned %s", domain.ErrUnavailable, resp.Status)
	}
	var r resolveResponse
	if err := json.Unmarshal(raw, &r); err != nil || r.CardID == "" {
		return domain.CardRef{}, fmt.Errorf("%w: card-service: bad resolve response", domain.ErrUnavailable)
	}
	return domain.CardRef{
		CardID: r.CardID, AccountID: r.AccountID, CustomerID: r.CustomerID,
		UserID: r.UserID, MaskedPAN: r.MaskedPAN, Status: r.Status,
	}, nil
}
