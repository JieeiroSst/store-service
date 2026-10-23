package http

import (
	"github.com/JIeeiroSst/movie-recommendation-service/pkg/authn"
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

	api := r.Group("/api/recommendations", authenticate(authn.FromEnv()))
	api.GET("/new", h.NewReleases)
	api.GET("/videos/:id", h.Video)
	api.GET("/videos/:id/similar", h.Similar)
	api.GET("/trending", h.Trending)

	users := api.Group("/users/:user_id", requireUser(), ownerParam("user_id"))
	users.GET("", h.ForUser)
	users.GET("/home", h.Home)
	users.GET("/continue-watching", h.ContinueWatching)
	users.GET("/history", h.History)
	users.DELETE("/history/:video_id", h.RemoveFromHistory)

	api.POST("/events", requireUser(), forceOwnerBody("user_id", false), h.RecordEvent)
	api.PUT("/videos/:id", requireAdmin(), h.UpsertVideo)
	api.DELETE("/videos/:id", requireAdmin(), h.DeleteVideo)
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
