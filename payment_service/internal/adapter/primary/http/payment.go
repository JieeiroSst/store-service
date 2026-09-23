package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type createPaymentRequest struct {
	Provider       string            `json:"provider" binding:"required"`
	Amount         int64             `json:"amount" binding:"required"`
	Currency       string            `json:"currency" binding:"required"`
	PayerEmail     string            `json:"payer_email"`
	Description    string            `json:"description"`
	Metadata       map[string]string `json:"metadata"`
	IdempotencyKey string            `json:"idempotency_key"`
}

func (r createPaymentRequest) toInput() port.CreatePaymentInput {
	return port.CreatePaymentInput{
		Provider:       model.Provider(r.Provider),
		Amount:         r.Amount,
		Currency:       r.Currency,
		PayerEmail:     r.PayerEmail,
		Description:    r.Description,
		Metadata:       r.Metadata,
		IdempotencyKey: r.IdempotencyKey,
	}
}

func (h *Handler) CreatePayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := req.toInput()
	if headerKey := c.GetHeader("Idempotency-Key"); headerKey != "" {
		input.IdempotencyKey = headerKey
	}

	payment, err := h.payments.CreatePayment(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, payment)
}

func (h *Handler) ListPayments(c *gin.Context) {
	in := port.ListPaymentsInput{
		Provider: model.Provider(c.Query("provider")),
		Status:   model.PaymentStatus(c.Query("status")),
	}
	if from := c.Query("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from: must be RFC3339"})
			return
		}
		in.From = t
	}
	if to := c.Query("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to: must be RFC3339"})
			return
		}
		in.To = t
	}
	if limit := c.Query("limit"); limit != "" {
		v, err := strconv.Atoi(limit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		in.Limit = v
	}
	if offset := c.Query("offset"); offset != "" {
		v, err := strconv.Atoi(offset)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
		in.Offset = v
	}

	result, err := h.payments.ListPayments(c.Request.Context(), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetPayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	payment, err := h.payments.GetPayment(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, payment)
}

type refundPaymentRequest struct {
	Amount int64 `json:"amount"`
}

func (h *Handler) RefundPayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req refundPaymentRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	payment, err := h.payments.RefundPayment(c.Request.Context(), id, req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, payment)
}

func (h *Handler) ListTransactions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	transactions, err := h.payments.ListTransactions(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, transactions)
}
