package http

import (
	"io"
	stdhttp "net/http"

	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	svc paymentapp.PaymentService
}

func NewPaymentHandler(svc paymentapp.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

type initiatePaymentRequest struct {
	OrderID        string `json:"order_id" binding:"required"`
	Amount         int64  `json:"amount" binding:"required"`
	Currency       string `json:"currency" binding:"required"`
	Provider       string `json:"provider" binding:"required"`
	ReturnURL      string `json:"return_url"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *PaymentHandler) Initiate(c *gin.Context) {
	var req initiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	orderID, err := shared.ParseOrderID(req.OrderID)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid order_id"}})
		return
	}
	out, err := h.svc.InitiatePayment(c.Request.Context(), paymentapp.InitiatePaymentInput{
		OrderID: orderID, Amount: shared.NewMoney(req.Amount, req.Currency), Provider: req.Provider,
		ReturnURL: req.ReturnURL, IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, out)
}

func (h *PaymentHandler) Webhook(c *gin.Context) {
	provider := c.Param("provider")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "unreadable body"}})
		return
	}
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		signature = c.Query("vnp_SecureHash")
	}

	if err := h.svc.HandleWebhook(c.Request.Context(), paymentapp.WebhookInput{Provider: provider, RawBody: body, Signature: signature}); err != nil {
		mapError(c, err)
		return
	}
	c.Status(stdhttp.StatusOK)
}
