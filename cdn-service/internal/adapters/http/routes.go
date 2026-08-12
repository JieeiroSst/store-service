package http

import (
	"net/http"

	"github.com/JIeeiroSst/cdn-service/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(cfg *config.Config, h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), Metrics(), RateLimit(20, 40))

	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	files := r.Group("/api/v1/files")
	{
		files.GET("", h.List)
		files.GET("/:id", h.Get)
		files.GET("/:id/download", h.Download)

		authed := files.Group("")
		authed.Use(APIKeyAuth(cfg))
		{
			authed.POST("/presign", h.Presign)
			authed.POST("/:id/confirm", h.Confirm)
			authed.DELETE("/:id", h.Delete)
		}
	}

	return r
}
