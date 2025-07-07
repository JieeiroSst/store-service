package handlers

import (
	"net/http"

	"github.com/JIeeiroSst/integrated-payment-service/internal/application/services"
	"github.com/JIeeiroSst/integrated-payment-service/pkg/utils"
	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) HandleMoMoWebhook(c *gin.Context) {
	var req services.MoMoWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid MoMo webhook request", err)
		return
	}

	if err := h.paymentService.HandleMoMoWebhook(c.Request.Context(), &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to handle MoMo webhook", err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *PaymentHandler) HandleVNPayWebhook(c *gin.Context) {
	var req services.VNPayWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid VNPay webhook request", err)
		return
	}

	if err := h.paymentService.HandleVNPayWebhook(c.Request.Context(), &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to handle VNPay webhook", err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *PaymentHandler) HandleZaloPayWebhook(c *gin.Context) {
	var req services.ZaloPayWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ZaloPay webhook request", err)
		return
	}

	if err := h.paymentService.HandleZaloPayWebhook(c.Request.Context(), &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to handle ZaloPay webhook", err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *PaymentHandler) HandleStripeWebhook(c *gin.Context) {
	var req services.StripeWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Stripe webhook request", err)
		return
	}

	if err := h.paymentService.HandleStripeWebhook(c.Request.Context(), &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to handle Stripe webhook", err)
		return
	}

	c.Status(http.StatusOK)
}
