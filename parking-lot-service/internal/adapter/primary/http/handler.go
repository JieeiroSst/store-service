package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	parking port.ParkingUsecase
	rates   port.RateUsecase
}

func NewHandler(parking port.ParkingUsecase, rates port.RateUsecase) *Handler {
	return &Handler{parking: parking, rates: rates}
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrNoSpotAvailable):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrTicketAlreadyClosed),
		errors.Is(err, port.ErrInvalidVehicleType),
		errors.Is(err, port.ErrInvalidPaymentMethod):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
