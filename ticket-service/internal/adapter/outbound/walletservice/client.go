package walletservice

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/ticket-service/config"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/httpx"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type Client struct{ c *httpx.Client }

var _ outbound.WalletGateway = (*Client)(nil)

func NewClient(cfg config.Config) *Client {
	return &Client{c: httpx.New("wallet-service", cfg.WalletServiceURL, cfg.PaymentServiceTimeout)}
}

type walletJSON struct {
	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id"`
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

func (w walletJSON) toDomain() domain.Wallet {
	uid, _ := strconv.ParseInt(w.UserID, 10, 64)
	return domain.Wallet{ID: w.WalletID, UserID: uid, Balance: w.Balance, Currency: w.Currency, Status: w.Status}
}

func (c *Client) GetByUser(ctx context.Context, userID int64) (domain.Wallet, error) {
	var w walletJSON
	st, msg, err := c.c.Do(ctx, http.MethodGet, "/api/v1/wallets/user/"+strconv.FormatInt(userID, 10), nil, nil, &w)
	if err != nil {
		return domain.Wallet{}, err
	}
	if !httpx.OK(st) {
		return domain.Wallet{}, mapStatus(st, msg)
	}
	return w.toDomain(), nil
}

func (c *Client) Transfer(ctx context.Context, p outbound.TransferParams) (string, error) {
	var out struct {
		TransferID string `json:"transfer_id"`
	}
	st, msg, err := c.c.Do(ctx, http.MethodPost, "/api/v1/transfers", nil, map[string]any{
		"sender_wallet_id": p.FromWalletID, "receiver_wallet_id": p.ToWalletID,
		"amount": p.Amount, "reference_id": p.ReferenceID, "description": p.Description,
	}, &out)
	if err != nil {
		return "", err
	}
	if !httpx.OK(st) {
		return "", mapStatus(st, msg)
	}
	if out.TransferID == "" {
		return "", fmt.Errorf("%w: wallet-service returned no transfer_id", domain.ErrUpstreamUnavailable)
	}
	return out.TransferID, nil
}

func (c *Client) ReverseTransfer(ctx context.Context, transferID, reason string) error {
	id := url.PathEscape(transferID)
	var t struct {
		Status string `json:"status"`
	}
	st, msg, err := c.c.Do(ctx, http.MethodGet, "/api/v1/transfers/"+id, nil, nil, &t)
	if err != nil {
		return err
	}
	if !httpx.OK(st) {
		return mapStatus(st, msg)
	}
	if t.Status == "REVERSED" {
		return nil
	}
	st, msg, err = c.c.Do(ctx, http.MethodPost, "/api/v1/transfers/"+id+"/reverse", nil, map[string]string{"reason": reason}, nil)
	if err != nil {
		return err
	}
	if !httpx.OK(st) {
		return mapStatus(st, msg)
	}
	return nil
}

func mapStatus(status int, msg string) error {
	switch status {
	case http.StatusNotFound:
		return domain.ErrNotFound
	case http.StatusConflict:
		return fmt.Errorf("%w: %s", domain.ErrConflict, msg)
	case http.StatusUnprocessableEntity:
		return domain.ErrInsufficientFunds
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", domain.ErrInvalid, msg)
	default:
		return fmt.Errorf("%w: wallet-service returned %d: %s", domain.ErrUpstreamUnavailable, status, msg)
	}
}
