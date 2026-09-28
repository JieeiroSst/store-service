package payment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

const (
	walletAttempts = 3
	walletBackoff  = 200 * time.Millisecond
	statusReversed = "REVERSED"
)

type Wallet struct {
	client *httpx.Client
}

func NewWallet(client *httpx.Client) *Wallet { return &Wallet{client: client} }

type walletTransaction struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

func (w *Wallet) Charge(ctx context.Context, req port.ChargeRequest) (string, error) {
	if req.WalletID == "" {
		return "", fmt.Errorf("wallet_id is required")
	}
	body := map[string]any{
		"amount":       req.AmountCents,
		"reference_id": req.PaymentID,
		"description":  "vending machine purchase " + req.PaymentID,
	}
	var txn walletTransaction
	err := httpx.WithRetry(ctx, walletAttempts, walletBackoff, func() error {
		return w.client.Do(ctx, http.MethodPost, "/api/v1/wallets/"+url.PathEscape(req.WalletID)+"/withdraw", body, &txn)
	})
	if err == nil {
		return txn.TransactionID, nil
	}
	if !httpx.Retryable(err) {
		return "", fmt.Errorf("wallet withdraw: %w", err)
	}
	lookupCtx := context.WithoutCancel(ctx)
	txID, found, lookupErr := w.Lookup(lookupCtx, port.LookupRequest{PaymentID: req.PaymentID, WalletID: req.WalletID})
	if lookupErr == nil && found {
		return txID, nil
	}
	return "", fmt.Errorf("%w: wallet withdraw: %v", domain.ErrPaymentPending, err)
}

func (w *Wallet) Lookup(ctx context.Context, req port.LookupRequest) (string, bool, error) {
	if req.WalletID == "" {
		return "", false, fmt.Errorf("wallet_id is required")
	}
	var txn walletTransaction
	path := "/api/v1/wallets/" + url.PathEscape(req.WalletID) + "/transactions/by-reference/" + url.PathEscape(req.PaymentID)
	err := httpx.WithRetry(ctx, walletAttempts, walletBackoff, func() error {
		return w.client.Do(ctx, http.MethodGet, path, nil, &txn)
	})
	switch {
	case err == nil:
		return txn.TransactionID, true, nil
	case httpx.IsStatus(err, http.StatusNotFound):
		return "", false, nil
	}
	return "", false, fmt.Errorf("wallet lookup: %w", err)
}

func (w *Wallet) Refund(ctx context.Context, req port.RefundRequest) error {
	path := "/api/v1/transactions/" + url.PathEscape(req.TransactionID)
	err := httpx.WithRetry(ctx, walletAttempts, walletBackoff, func() error {
		return w.client.Do(ctx, http.MethodPost, path+"/reverse", map[string]string{"reason": req.Reason}, nil)
	})
	if err == nil || !httpx.IsStatus(err, http.StatusBadRequest) {
		return wrap("wallet reverse", err)
	}
	var txn walletTransaction
	if getErr := w.client.Do(ctx, http.MethodGet, path, nil, &txn); getErr == nil && txn.Status == statusReversed {
		return nil
	}
	return wrap("wallet reverse", err)
}

func wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}
