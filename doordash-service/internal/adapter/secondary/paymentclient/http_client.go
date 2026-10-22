package paymentclient

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/JIeeiroSst/doordash-service/config"
	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

var errCardNotOwned = errors.New("payment card does not belong to the customer")

type walletPaymentClient struct {
	client     *http.Client
	baseURL    string
	merchantID string
	currency   string
}

func NewPaymentClient(cfg *config.Config) port.PaymentClient {
	return &walletPaymentClient{
		client:     &http.Client{Timeout: cfg.PaymentService.TimeoutDuration()},
		baseURL:    cfg.PaymentService.BaseURL,
		merchantID: cfg.Wallet.MerchantID,
		currency:   cfg.Wallet.Currency,
	}
}

type walletCard struct {
	WalletID string `json:"wallet_id"`
}

type walletList struct {
	Items []struct {
		ID string `json:"id"`
	} `json:"items"`
}

type authorizeRequest struct {
	IdempotencyKey string            `json:"idempotency_key"`
	CardID         string            `json:"card_id"`
	MerchantID     string            `json:"merchant_id"`
	Amount         string            `json:"amount"`
	Currency       string            `json:"currency"`
	Description    string            `json:"description"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type walletTransaction struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (c *walletPaymentClient) AuthorizePayment(ctx context.Context, in port.AuthorizePaymentInput) (*model.PaymentAuthorization, error) {
	if err := c.ensureCardOwner(ctx, in.CustomerID, in.PaymentMethodID); err != nil {
		if errors.Is(err, errCardNotOwned) {
			return &model.PaymentAuthorization{Status: model.PaymentStatusFailed}, nil
		}
		return nil, err
	}

	var tx walletTransaction
	status, err := c.do(ctx, http.MethodPost, "/api/v1/transactions/authorize", authorizeRequest{
		IdempotencyKey: idempotencyKey(in),
		CardID:         in.PaymentMethodID,
		MerchantID:     c.merchantID,
		Amount:         strconv.FormatFloat(in.Amount, 'f', 2, 64),
		Currency:       c.currency,
		Description:    "DoorDash order",
		Metadata:       map[string]string{"customer_id": in.CustomerID},
	}, &tx)
	if err != nil {
		return nil, err
	}
	if status >= 400 || tx.Status != "AUTHORIZED" {
		return &model.PaymentAuthorization{PaymentID: tx.ID, Status: model.PaymentStatusFailed}, nil
	}
	return &model.PaymentAuthorization{PaymentID: tx.ID, Status: model.PaymentStatusAuthorized}, nil
}

func (c *walletPaymentClient) CapturePayment(ctx context.Context, paymentID string) error {
	return c.settle(ctx, paymentID, "capture")
}

func (c *walletPaymentClient) VoidPayment(ctx context.Context, paymentID string) error {
	return c.settle(ctx, paymentID, "void")
}

func (c *walletPaymentClient) settle(ctx context.Context, paymentID, action string) error {
	status, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/transactions/%s/%s", url.PathEscape(paymentID), action), struct{}{}, nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("wallet-service %s %s returned status %d", action, paymentID, status)
	}
	return nil
}

func (c *walletPaymentClient) ensureCardOwner(ctx context.Context, customerID, cardID string) error {
	if customerID == "" || cardID == "" {
		return errCardNotOwned
	}
	var card walletCard
	status, err := c.do(ctx, http.MethodGet, "/api/v1/cards/"+url.PathEscape(cardID), nil, &card)
	if err != nil {
		return err
	}
	if status >= 400 {
		return errCardNotOwned
	}
	var wallets walletList
	status, err = c.do(ctx, http.MethodGet, "/api/v1/users/"+url.PathEscape(customerID)+"/wallets", nil, &wallets)
	if err != nil {
		return err
	}
	if status >= 400 {
		return errCardNotOwned
	}
	for _, w := range wallets.Items {
		if w.ID == card.WalletID {
			return nil
		}
	}
	return errCardNotOwned
}

func (c *walletPaymentClient) do(ctx context.Context, method, path string, body, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("encode wallet-service request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("call wallet-service: %w", err)
	}
	defer resp.Body.Close()
	if out != nil && resp.StatusCode < 400 {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, fmt.Errorf("decode wallet-service response: %w", err)
		}
	}
	return resp.StatusCode, nil
}

func idempotencyKey(in port.AuthorizePaymentInput) string {
	if in.OrderID > 0 {
		return fmt.Sprintf("doordash-order-%d", in.OrderID)
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("doordash-%s-%d", in.CustomerID, time.Now().UnixNano())
	}
	return "doordash-" + hex.EncodeToString(b)
}
