package http

import (
	"net/http"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(),
		gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/healthz", "/readyz"}}),
		cors(cfg.Server.AllowedOrigins))

	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/readyz", h.Readyz)

	api := r.Group("/api/recommendations")
	api.GET("/users/:user_id", h.ForUser)
	api.GET("/users/:user_id/home", h.Home)
	api.GET("/users/:user_id/continue-watching", h.ContinueWatching)
	api.GET("/users/:user_id/history", h.History)
	api.DELETE("/users/:user_id/history/:video_id", h.RemoveFromHistory)
	api.GET("/new", h.NewReleases)
	api.GET("/videos/:id/similar", h.Similar)
	api.GET("/trending", h.Trending)
	api.POST("/events", h.RecordEvent)
	api.PUT("/videos/:id", h.UpsertVideo)
	api.DELETE("/videos/:id", h.DeleteVideo)
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
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
