package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	notification port.NotificationUsecase
	device       port.UserDeviceUsecase
	campaign     port.CampaignUsecase
	audit        port.AuditUsecase
}

func NewHandler(notification port.NotificationUsecase, device port.UserDeviceUsecase, campaign port.CampaignUsecase, audit port.AuditUsecase) *Handler {
	return &Handler{
		notification: notification,
		device:       device,
		campaign:     campaign,
		audit:        audit,
	}
}

const headerRequestedBy = "X-Requested-By"

func requestedBy(c *gin.Context) string {
	v := strings.TrimSpace(c.GetHeader(headerRequestedBy))
	if len(v) > 128 {
		v = v[:128]
	}
	return v
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
