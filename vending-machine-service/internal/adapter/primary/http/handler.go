package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/gin-gonic/gin"
)

const readyTimeout = 2 * time.Second

type Handler struct {
	machines  port.MachineService
	catalog   port.CatalogService
	inventory port.InventoryService
	vending   port.VendingService
	health    port.HealthChecker
}

func NewHandler(
	machines port.MachineService,
	catalog port.CatalogService,
	inventory port.InventoryService,
	vending port.VendingService,
	health port.HealthChecker,
) *Handler {
	return &Handler{machines: machines, catalog: catalog, inventory: inventory, vending: vending, health: health}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), readyTimeout)
	defer cancel()
	if err := h.health.Ping(ctx); err != nil {
		log.Printf("readiness: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func bind(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body: " + err.Error()})
		return false
	}
	return true
}

func statusOf(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrPaymentFailed):
		return http.StatusPaymentRequired
	case errors.Is(err, domain.ErrCouponRejected):
		return http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrUpstream):
		return http.StatusBadGateway
	case errors.Is(err, domain.ErrConflict),
		errors.Is(err, domain.ErrMachineUnavailable),
		errors.Is(err, domain.ErrSessionNotActive),
		errors.Is(err, domain.ErrProductInactive),
		errors.Is(err, domain.ErrOutOfStock),
		errors.Is(err, domain.ErrReservationClosed),
		errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrPriceChanged):
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

func writeError(c *gin.Context, err error) {
	var pending *domain.PendingPaymentError
	if errors.As(err, &pending) {
		c.JSON(http.StatusAccepted, gin.H{
			"status":     "pending",
			"payment_id": pending.PaymentID,
			"message":    pending.Error(),
			"status_url": "/api/v1/payments/" + pending.PaymentID,
		})
		return
	}
	status := statusOf(err)
	if status >= http.StatusInternalServerError {
		log.Printf("%s %s: %v", c.Request.Method, c.FullPath(), err)
	}
	if status == http.StatusInternalServerError {
		c.JSON(status, gin.H{"error": "internal error"})
		return
	}
	c.JSON(status, gin.H{"error": err.Error()})
}
