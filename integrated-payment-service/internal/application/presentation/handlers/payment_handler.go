package handlers

import (
	"net/http"

	"github.com/JIeeiroSst/integrated-payment-service/internal/application/services"
	"github.com/JIeeiroSst/integrated-payment-service/pkg/utils"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req services.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	response, err := h.paymentService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create payment", err)
		return
	}

	utils.SuccessResponse(c, response)
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	paymentID := c.Param("id")

	response, err := h.paymentService.ProcessPayment(c.Request.Context(), paymentID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process payment", err)
		return
	}

	utils.SuccessResponse(c, response)
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentID := c.Param("id")

	var req struct {
		Amount float64 `json:"amount" validate:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	response, err := h.paymentService.RefundPayment(c.Request.Context(), paymentID, req.Amount)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to refund payment", err)
		return
	}

	utils.SuccessResponse(c, response)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")

	response, err := h.paymentService.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get payment", err)
		return
	}

	utils.SuccessResponse(c, response)
}

func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	paymentID := c.Param("id")

	response, err := h.paymentService.GetPaymentStatus(c.Request.Context(), paymentID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get payment status", err)
		return
	}

	utils.SuccessResponse(c, response)
}
