package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/payment-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	payments port.PaymentUsecase
}

func NewHandler(payments port.PaymentUsecase) *Handler {
	return &Handler{payments: payments}
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrInvalidProvider),
		errors.Is(err, port.ErrUnsupportedProvider),
		errors.Is(err, port.ErrInvalidAmount),
		errors.Is(err, port.ErrInvalidCurrency),
		errors.Is(err, port.ErrInvalidRefundAmount),
		errors.Is(err, port.ErrInvalidWebhook):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrPaymentNotRefundable),
		errors.Is(err, port.ErrRefundNotSupported):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrGatewayRequestFailed):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
