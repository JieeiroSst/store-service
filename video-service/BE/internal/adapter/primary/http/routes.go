package http

import (
	"net/http"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/healthz"}}), cors(cfg.Server.AllowedOrigins))

	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	api := r.Group("/api/videos")
	api.GET("", h.List)
	api.POST("", h.Upload)
	api.GET("/:id", h.Get)
	api.GET("/:id/related", h.Related)
	api.POST("/:id/view", h.View)
	api.GET("/:id/thumbnail", h.Thumbnail)
	api.GET("/:id/hls/*path", h.HLS)
	api.GET("/:id/stream", h.Stream)
	api.HEAD("/:id/stream", h.Stream)
	api.DELETE("/:id", h.Delete)
	return r
}

func cors(origins []string) gin.HandlerFunc {
	allowAll := len(origins) == 1 && origins[0] == "*"
	allowed := map[string]bool{}
	for _, o := range origins {
		allowed[o] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		switch {
		case allowAll:
			c.Header("Access-Control-Allow-Origin", "*")
		case allowed[origin]:
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, HEAD, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Range, If-Range, If-None-Match")
		c.Header("Access-Control-Expose-Headers", "Content-Range, Content-Length, Accept-Ranges, ETag")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
