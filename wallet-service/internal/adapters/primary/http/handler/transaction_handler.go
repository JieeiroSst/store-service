package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)

type TransactionHandler struct {
	txUC ports.TransactionUseCase
}

func NewTransactionHandler(txUC ports.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{txUC: txUC}
}

// POST /transactions/authorize
// Implements VISA Authorization Flow steps 1-4.3
func (h *TransactionHandler) Authorize(c *gin.Context) {
	var req struct {
		IdempotencyKey string `json:"idempotency_key" binding:"required"`
		CardID         string `json:"card_id" binding:"required,uuid"`
		MerchantID     string `json:"merchant_id" binding:"required,uuid"`
		Amount         string `json:"amount" binding:"required"`
		Currency       string `json:"currency" binding:"required"`
		Description    string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cardID, _ := uuid.Parse(req.CardID)
	merchantID, _ := uuid.Parse(req.MerchantID)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}

	tx, err := h.txUC.Authorize(c.Request.Context(), ports.AuthorizeRequest{
		IdempotencyKey: req.IdempotencyKey,
		CardID:         cardID,
		MerchantID:     merchantID,
		Amount:         amount,
		Currency:       domain.Currency(req.Currency),
		Description:    req.Description,
	})
	if err != nil {
		switch err {
		case domain.ErrInsufficientFunds:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "code": "INSUFFICIENT_FUNDS"})
		case domain.ErrCardExpired, domain.ErrCardBlocked:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "code": "CARD_DECLINED"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, tx)
}

// POST /transactions/:id/capture
// Implements Settlement Flow step 1 (merchant end-of-day capture)
func (h *TransactionHandler) Capture(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	tx, err := h.txUC.Capture(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tx)
}

// POST /transactions/:id/void
func (h *TransactionHandler) Void(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	tx, err := h.txUC.Void(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tx)
}

// GET /transactions/:id
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}
	tx, err := h.txUC.GetTransaction(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tx)
}

// POST /settlements/batches
// Triggers Settlement Flow steps 1-2 (merchant batches transactions)
func (h *TransactionHandler) CreateSettlementBatch(c *gin.Context) {
	var req struct {
		MerchantID string `json:"merchant_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	merchantID, _ := uuid.Parse(req.MerchantID)
	batch, err := h.txUC.CreateSettlementBatch(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, batch)
}

// POST /settlements/batches/:batch_id/clear
// Settlement Flow step 3: card network clearing with netting
func (h *TransactionHandler) ProcessClearing(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("batch_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch id"})
		return
	}
	records, err := h.txUC.ProcessClearing(c.Request.Context(), batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"clearing_records": records, "count": len(records)})
}

// POST /settlements/batches/:batch_id/settle
// Settlement Flow steps 4-5: issuer confirms, money transferred to acquirer then merchant
func (h *TransactionHandler) ProcessSettlement(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("batch_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch id"})
		return
	}
	if err := h.txUC.ProcessSettlement(c.Request.Context(), batchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "settled"})
}
