package walletservice

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/JIeerioSst/rent-house-service/config"
	"github.com/JIeerioSst/rent-house-service/internal/adapter/outbound/httpx"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
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

func (c *Client) Get(ctx context.Context, walletID string) (domain.Wallet, error) {
	var w walletJSON
	st, msg, err := c.c.Do(ctx, http.MethodGet, "/api/v1/wallets/"+url.PathEscape(walletID), nil, nil, &w)
	if err != nil {
		return domain.Wallet{}, err
	}
	if !httpx.OK(st) {
		return domain.Wallet{}, mapStatus(st, msg)
	}
	return w.toDomain(), nil
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

func (c *Client) Create(ctx context.Context, userID int64, currency string) (domain.Wallet, error) {
	var w walletJSON
	st, msg, err := c.c.Do(ctx, http.MethodPost, "/api/v1/wallets", nil,
		map[string]string{"user_id": strconv.FormatInt(userID, 10), "currency": currency}, &w)
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

func (c *Client) Transactions(ctx context.Context, walletID string, limit, offset int) ([]domain.WalletTxn, error) {
	q := url.Values{"limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(offset)}}
	var raw []struct {
		ID          string    `json:"transaction_id"`
		Type        string    `json:"type"`
		Amount      int64     `json:"amount"`
		Currency    string    `json:"currency"`
		Status      string    `json:"status"`
		ReferenceID string    `json:"reference_id"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"created_at"`
	}
	st, msg, err := c.c.Do(ctx, http.MethodGet, "/api/v1/wallets/"+url.PathEscape(walletID)+"/transactions?"+q.Encode(), nil, nil, &raw)
	if err != nil {
		return nil, err
	}
	if !httpx.OK(st) {
		return nil, mapStatus(st, msg)
	}
	out := make([]domain.WalletTxn, len(raw))
	for i, t := range raw {
		out[i] = domain.WalletTxn{ID: t.ID, Type: t.Type, Amount: t.Amount, Currency: t.Currency,
			Status: t.Status, ReferenceID: t.ReferenceID, Description: t.Description, CreatedAt: t.CreatedAt}
	}
	return out, nil
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
