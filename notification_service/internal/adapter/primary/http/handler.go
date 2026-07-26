package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	notification port.NotificationUsecase
	device       port.UserDeviceUsecase
}

func NewHandler(notification port.NotificationUsecase, device port.UserDeviceUsecase) *Handler {
	return &Handler{
		notification: notification,
		device:       device,
	}
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrInvalidRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
