package http

import (
	"net/http"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

type HealthHandler struct {
	health *app.HealthService
}

func NewHealthHandler(health *app.HealthService) *HealthHandler {
	return &HealthHandler{health}
}

func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"uptime": int(time.Since(startTime).Seconds()),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	result := h.health.Ready(c.Request.Context())
	checks := gin.H{"database": result.Database}
	if result.Redis != "" {
		checks["redis"] = result.Redis
	}
	status := "ok"
	code := http.StatusOK
	if !result.Ready {
		status = "error"
		code = http.StatusServiceUnavailable
	}
	c.JSON(code, gin.H{"status": status, "checks": checks})
}
