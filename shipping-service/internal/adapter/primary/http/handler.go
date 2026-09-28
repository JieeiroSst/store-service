package http

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

const (
	headerAPIKey = "X-API-Key"
	clientKey    = "client"
)

type Handler struct {
	shipments  port.ShipmentUsecase
	webhooks   port.WebhookUsecase
	warehouses port.WarehouseUsecase
	locations  port.LocationUsecase
	apiKeys    map[string]string
}

func NewHandler(
	shipments port.ShipmentUsecase,
	webhooks port.WebhookUsecase,
	warehouses port.WarehouseUsecase,
	locations port.LocationUsecase,
	cfg *config.Config,
) *Handler {
	return &Handler{
		shipments:  shipments,
		webhooks:   webhooks,
		warehouses: warehouses,
		locations:  locations,
		apiKeys:    cfg.Auth.Keys(),
	}
}

func (h *Handler) authenticate(c *gin.Context) {
	presented := c.GetHeader(headerAPIKey)
	for key, client := range h.apiKeys {
		if presented != "" && subtle.ConstantTimeCompare([]byte(presented), []byte(key)) == 1 {
			c.Set(clientKey, client)
			c.Next()
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid " + headerAPIKey})
}

func client(c *gin.Context) string { return c.GetString(clientKey) }

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, port.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, port.ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, port.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, port.ErrConflict), errors.Is(err, port.ErrInvalidTransition):
		status = http.StatusConflict
	case errors.Is(err, port.ErrNoRoute), errors.Is(err, port.ErrCarrierRejected):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, port.ErrCarrierUnavailable):
		status = http.StatusBadGateway
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func bind(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, name string) int {
	v, _ := strconv.Atoi(c.Query(name))
	return v
}
