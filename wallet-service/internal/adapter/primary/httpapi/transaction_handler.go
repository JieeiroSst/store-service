package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func (a *API) authorize(c *gin.Context) {
	var req struct {
		IdempotencyKey string            `json:"idempotency_key"`
		CardID         string            `json:"card_id"`
		MerchantID     string            `json:"merchant_id"`
		Amount         decimal.Decimal   `json:"amount"`
		Currency       string            `json:"currency" binding:"required"`
		Description    string            `json:"description"`
		Metadata       map[string]string `json:"metadata"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.AuthorizeCommand{
		IdempotencyKey: idempotencyKey(c, req.IdempotencyKey),
		CardID:         p.required("card_id", req.CardID),
		MerchantID:     p.required("merchant_id", req.MerchantID),
		Amount:         req.Amount,
		Currency:       model.Currency(req.Currency),
		Description:    req.Description,
		Metadata:       req.Metadata,
	}
	if !p.ok(c) {
		return
	}
	tx, err := a.CardPayments.Authorize(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, tx)
}

func (a *API) capture(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	tx, err := a.CardPayments.Capture(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (a *API) void(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	tx, err := a.CardPayments.Void(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (a *API) getTransaction(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	tx, err := a.History.Transaction(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (a *API) createSettlementBatch(c *gin.Context) {
	var req struct {
		MerchantID string `json:"merchant_id"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	merchantID := p.required("merchant_id", req.MerchantID)
	if !p.ok(c) {
		return
	}
	batch, err := a.Settlements.CreateBatch(c.Request.Context(), merchantID)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, batch)
}

func (a *API) clearBatch(c *gin.Context) {
	id, ok := pathID(c, "batch_id")
	if !ok {
		return
	}
	records, err := a.Settlements.Clear(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"clearing_records": records, "count": len(records)})
}

func (a *API) settleBatch(c *gin.Context) {
	id, ok := pathID(c, "batch_id")
	if !ok {
		return
	}
	if err := a.Settlements.Settle(c.Request.Context(), id); err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "settled"})
}
