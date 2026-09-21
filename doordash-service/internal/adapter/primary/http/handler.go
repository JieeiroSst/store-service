package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	orders      port.OrderUsecase
	tracking    port.TrackingUsecase
	assignments port.DriverAssignmentUsecase
}

func NewHandler(orders port.OrderUsecase, tracking port.TrackingUsecase, assignments port.DriverAssignmentUsecase) *Handler {
	return &Handler{orders: orders, tracking: tracking, assignments: assignments}
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrEmptyOrder),
		errors.Is(err, port.ErrInvalidStatusTransition):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrOrderAlreadyTerminal),
		errors.Is(err, port.ErrRestaurantInactive),
		errors.Is(err, port.ErrMenuItemUnavailable),
		errors.Is(err, port.ErrPaymentAuthorizationFailed),
		errors.Is(err, port.ErrDriverUnavailable),
		errors.Is(err, port.ErrAssignmentNotPending):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
