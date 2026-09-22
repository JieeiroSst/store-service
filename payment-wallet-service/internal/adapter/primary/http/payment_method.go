package http

import (
	"net/http"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

type addPaymentMethodRequest struct {
	UserID        string                  `json:"user_id" binding:"required"`
	Type          model.PaymentMethodType `json:"type" binding:"required"`
	Provider      string                  `json:"provider" binding:"required"`
	AccountNumber string                  `json:"account_number" binding:"required"`
}

func (h *Handler) AddPaymentMethod(c *gin.Context) {
	var req addPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	method, err := h.paymentMethods.AddPaymentMethod(c.Request.Context(), req.UserID, req.Type, req.Provider, req.AccountNumber)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, method)
}

func (h *Handler) ListPaymentMethods(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	methods, err := h.paymentMethods.ListPaymentMethods(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, methods)
}

func (h *Handler) RemovePaymentMethod(c *gin.Context) {
	if err := h.paymentMethods.RemovePaymentMethod(c.Request.Context(), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) SetDefaultPaymentMethod(c *gin.Context) {
	method, err := h.paymentMethods.SetDefaultPaymentMethod(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, method)
}
