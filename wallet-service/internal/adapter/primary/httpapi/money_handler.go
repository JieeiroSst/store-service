package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func (a *API) topUp(c *gin.Context) {
	walletID, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		UserID         string          `json:"user_id"`
		BankAccountID  string          `json:"bank_account_id"`
		Amount         decimal.Decimal `json:"amount"`
		Description    string          `json:"description"`
		IdempotencyKey string          `json:"idempotency_key"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.TopUpCommand{
		IdempotencyKey: idempotencyKey(c, req.IdempotencyKey),
		UserID:         p.requiredUser("user_id", req.UserID),
		WalletID:       walletID,
		BankAccountID:  p.required("bank_account_id", req.BankAccountID),
		Amount:         req.Amount,
		Description:    req.Description,
	}
	if !p.ok(c) {
		return
	}
	tx, err := a.Funding.TopUp(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, tx)
}

func (a *API) withdraw(c *gin.Context) {
	walletID, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		UserID         string          `json:"user_id"`
		BankAccountID  string          `json:"bank_account_id"`
		Amount         decimal.Decimal `json:"amount"`
		Pin            string          `json:"pin"`
		Description    string          `json:"description"`
		IdempotencyKey string          `json:"idempotency_key"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.WithdrawCommand{
		IdempotencyKey: idempotencyKey(c, req.IdempotencyKey),
		UserID:         p.requiredUser("user_id", req.UserID),
		WalletID:       walletID,
		BankAccountID:  p.required("bank_account_id", req.BankAccountID),
		Amount:         req.Amount,
		Pin:            req.Pin,
		Description:    req.Description,
	}
	if !p.ok(c) {
		return
	}
	tx, err := a.Funding.Withdraw(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, tx)
}

func (a *API) transfer(c *gin.Context) {
	var req struct {
		UserID         string          `json:"user_id"`
		FromWalletID   string          `json:"from_wallet_id"`
		ToWalletID     string          `json:"to_wallet_id"`
		ToUserID       string          `json:"to_user_id"`
		Amount         decimal.Decimal `json:"amount"`
		Pin            string          `json:"pin"`
		Description    string          `json:"description"`
		IdempotencyKey string          `json:"idempotency_key"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.TransferCommand{
		IdempotencyKey: idempotencyKey(c, req.IdempotencyKey),
		UserID:         p.requiredUser("user_id", req.UserID),
		FromWalletID:   p.required("from_wallet_id", req.FromWalletID),
		ToWalletID:     p.optional("to_wallet_id", req.ToWalletID),
		ToUserID:       p.optionalUser("to_user_id", req.ToUserID),
		Amount:         req.Amount,
		Pin:            req.Pin,
		Description:    req.Description,
	}
	if !p.ok(c) {
		return
	}
	tx, err := a.Transfers.Transfer(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, tx)
}

func (a *API) pay(c *gin.Context) {
	var req struct {
		UserID         string            `json:"user_id"`
		WalletID       string            `json:"wallet_id"`
		MerchantID     string            `json:"merchant_id"`
		Amount         decimal.Decimal   `json:"amount"`
		Pin            string            `json:"pin"`
		Description    string            `json:"description"`
		Metadata       map[string]string `json:"metadata"`
		IdempotencyKey string            `json:"idempotency_key"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.PayCommand{
		IdempotencyKey: idempotencyKey(c, req.IdempotencyKey),
		UserID:         p.requiredUser("user_id", req.UserID),
		WalletID:       p.required("wallet_id", req.WalletID),
		MerchantID:     p.required("merchant_id", req.MerchantID),
		Amount:         req.Amount,
		Pin:            req.Pin,
		Description:    req.Description,
		Metadata:       req.Metadata,
	}
	if !p.ok(c) {
		return
	}
	tx, err := a.Payments.Pay(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, tx)
}

func (a *API) refund(c *gin.Context) {
	txID, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Amount         decimal.Decimal `json:"amount"`
		Reason         string          `json:"reason"`
		IdempotencyKey string          `json:"idempotency_key"`
	}
	if !bind(c, &req) {
		return
	}
	tx, err := a.Payments.Refund(c.Request.Context(), port.RefundCommand{
		IdempotencyKey: idempotencyKey(c, req.IdempotencyKey),
		TransactionID:  txID,
		Amount:         req.Amount,
		Reason:         req.Reason,
	})
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, tx)
}
