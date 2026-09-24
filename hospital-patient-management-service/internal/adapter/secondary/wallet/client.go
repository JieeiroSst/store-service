package wallet

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/httpx"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct {
	http       *httpx.Client
	hospitalID string
	minorUnit  int64
	userPrefix string
}

func New(cfg *config.Config) *Client {
	w := cfg.Wallet
	return &Client{
		http:       httpx.New(w.BaseURL, w.Timeout),
		hospitalID: w.HospitalWalletID,
		minorUnit:  w.MinorUnit,
		userPrefix: w.UserPrefix,
	}
}

func (c *Client) Enabled() bool { return c.http.Enabled() && c.hospitalID != "" }

// Charge moves amount from the patient's wallet to the hospital's. The
// reference makes it idempotent on the wallet side.
func (c *Client) Charge(ctx context.Context, in port.ChargeInput) (string, error) {
	userID := in.UserID
	if userID == "" {
		userID = fmt.Sprintf("%s%d", c.userPrefix, in.PatientID)
	}

	var w struct {
		WalletID string `json:"wallet_id"`
	}
	if err := c.http.Do(ctx, http.MethodGet, "/api/v1/wallets/user/"+url.PathEscape(userID), nil, &w); err != nil {
		if httpx.IsStatus(err, http.StatusNotFound) {
			return "", model.PaymentFailed("patient %d has no wallet", in.PatientID)
		}
		return "", c.wrap(err)
	}

	var t struct {
		TransferID string `json:"transfer_id"`
	}
	err := c.http.Do(ctx, http.MethodPost, "/api/v1/transfers", map[string]any{
		"sender_wallet_id":   w.WalletID,
		"receiver_wallet_id": c.hospitalID,
		"amount":             c.minor(in.Amount),
		"reference_id":       in.Reference,
		"description":        in.Description,
	}, &t)
	if err != nil {
		if httpx.IsStatus(err, http.StatusUnprocessableEntity) {
			return "", model.PaymentFailed("insufficient wallet balance")
		}
		return "", c.wrap(err)
	}
	if t.TransferID == "" {
		return "", fmt.Errorf("%w: wallet-service returned no transfer id", model.ErrUpstream)
	}
	return t.TransferID, nil
}

// Refund reverses a transfer. A transfer that is already reversed counts as
// refunded, so a retry after a lost answer or an interrupted reopen succeeds.
func (c *Client) Refund(ctx context.Context, transferID, reason string) error {
	var cur struct {
		Status string `json:"status"`
	}
	path := "/api/v1/transfers/" + url.PathEscape(transferID)
	if err := c.http.Do(ctx, http.MethodGet, path, nil, &cur); err == nil && strings.EqualFold(cur.Status, "REVERSED") {
		return nil
	}
	err := c.http.Do(ctx, http.MethodPost, "/api/v1/transfers/"+url.PathEscape(transferID)+"/reverse",
		map[string]string{"reason": reason}, nil)
	if err != nil {
		return c.wrap(err)
	}
	return nil
}

func (c *Client) minor(amount float64) int64 {
	return int64(math.Round(amount * float64(c.minorUnit)))
}

// wrap sorts wallet failures into "we could not ask" and "the wallet said no".
func (c *Client) wrap(err error) error {
	if se, ok := httpx.AsStatus(err); ok && se.Status >= 400 && se.Status < 500 {
		return model.PaymentFailed("wallet-service: %s", se.Message)
	}
	return fmt.Errorf("%w: wallet-service: %v", model.ErrUpstream, err)
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.WalletGateway)))),
)
