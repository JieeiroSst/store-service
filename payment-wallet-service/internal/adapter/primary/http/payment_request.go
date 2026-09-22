package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type createPaymentRequestRequest struct {
	RequesterWalletID string  `json:"requester_wallet_id" binding:"required"`
	PayerWalletID     *string `json:"payer_wallet_id"`
	Amount            int64   `json:"amount" binding:"required"`
	Description       string  `json:"description"`
	ExpiresInSeconds  int64   `json:"expires_in_seconds"`
}

func (h *Handler) CreatePaymentRequest(c *gin.Context) {
	var req createPaymentRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request, err := h.paymentRequests.CreatePaymentRequest(
		c.Request.Context(),
		req.RequesterWalletID,
		req.PayerWalletID,
		req.Amount,
		req.Description,
		time.Duration(req.ExpiresInSeconds)*time.Second,
	)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, request)
}

func (h *Handler) GetPaymentRequest(c *gin.Context) {
	request, err := h.paymentRequests.GetPaymentRequest(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, request)
}

func (h *Handler) GetPaymentRequestQR(c *gin.Context) {
	request, err := h.paymentRequests.GetPaymentRequest(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	payload := fmt.Sprintf("paywallet://pay?request_id=%s&amount=%d&currency=%s", request.PaymentRequestID, request.Amount, request.Currency)
	c.JSON(http.StatusOK, gin.H{"payment_request_id": request.PaymentRequestID, "payload": payload})
}

func (h *Handler) ListPaymentRequests(c *gin.Context) {
	walletID := c.Query("wallet_id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet_id is required"})
		return
	}

	requests, err := h.paymentRequests.ListPaymentRequests(c.Request.Context(), walletID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, requests)
}

type payPaymentRequestRequest struct {
	PayerWalletID string `json:"payer_wallet_id" binding:"required"`
}

func (h *Handler) PayPaymentRequest(c *gin.Context) {
	var req payPaymentRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.paymentRequests.Pay(c.Request.Context(), c.Param("id"), req.PayerWalletID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, transfer)
}

func (h *Handler) CancelPaymentRequest(c *gin.Context) {
	if err := h.paymentRequests.Cancel(c.Request.Context(), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
