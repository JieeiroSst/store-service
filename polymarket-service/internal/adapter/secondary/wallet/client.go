package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/JIeeiroSst/polymarket-service/config"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.Wallet.BaseURL, "/"),
		http:    &http.Client{Timeout: cfg.Wallet.TimeoutDuration()},
	}
}

type walletDTO struct {
	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id"`
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

type transferDTO struct {
	TransferID string `json:"transfer_id"`
}

func (c *Client) GetWalletByUser(ctx context.Context, userID string) (*port.Wallet, error) {
	var w walletDTO
	if err := c.do(ctx, http.MethodGet, "/api/v1/wallets/user/"+url.PathEscape(userID), nil, &w); err != nil {
		return nil, err
	}
	return &port.Wallet{ID: w.WalletID, UserID: w.UserID, Balance: w.Balance, Currency: w.Currency, Status: w.Status}, nil
}

func (c *Client) Transfer(ctx context.Context, in port.TransferInput) (string, error) {
	var t transferDTO
	err := c.do(ctx, http.MethodPost, "/api/v1/transfers", map[string]any{
		"sender_wallet_id":   in.SenderWalletID,
		"receiver_wallet_id": in.ReceiverWalletID,
		"amount":             in.Amount,
		"reference_id":       in.ReferenceID,
		"description":        in.Description,
	}, &t)
	if err != nil {
		return "", err
	}
	return t.TransferID, nil
}

func (c *Client) ReverseTransfer(ctx context.Context, transferID, reason string) error {
	return c.do(ctx, http.MethodPost, "/api/v1/transfers/"+url.PathEscape(transferID)+"/reverse", map[string]any{"reason": reason}, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
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
		return fmt.Errorf("%w: %v", port.ErrWalletUnavailable, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil {
			return nil
		}
		return json.Unmarshal(raw, out)
	}
	return statusError(resp.StatusCode, raw)
}

func statusError(status int, raw []byte) error {
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(raw, &payload)
	msg := payload.Error
	if msg == "" {
		msg = strings.TrimSpace(string(raw))
	}

	switch {
	case status == http.StatusNotFound:
		return port.ErrWalletNotFound
	case status == http.StatusUnprocessableEntity:
		return port.ErrInsufficientFunds
	case status >= 400 && status < 500:
		return fmt.Errorf("%w: %s", port.ErrWalletRejected, msg)
	default:
		return fmt.Errorf("%w: status %d: %s", port.ErrWalletUnavailable, status, msg)
	}
}
