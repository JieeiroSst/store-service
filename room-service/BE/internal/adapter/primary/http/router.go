package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/room-service/config"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz"},
		Formatter: logLine,
	}), cors(cfg.Server.AllowedOrigins))

	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	api := r.Group("/api")
	api.POST("/auth/login", h.Login)
	api.POST("/auth/refresh", h.Refresh)

	authed := api.Group("", h.Authenticate())
	authed.GET("/me", h.Me)
	authed.GET("/rooms", h.ListRooms)
	authed.POST("/rooms", h.CreateRoom)
	authed.GET("/rooms/:id", h.GetRoom)
	authed.GET("/rooms/:id/members", h.ListMembers)
	authed.POST("/rooms/:id/members", h.AddMember)
	authed.DELETE("/rooms/:id/members/:username", h.RemoveMember)
	authed.GET("/rooms/:id/messages", h.Messages)
	authed.GET("/rooms/:id/ws", h.Connect)
	return r
}

func logLine(p gin.LogFormatterParams) string {
	path, _, _ := strings.Cut(p.Path, "?")
	return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %q\n",
		p.TimeStamp.Format("2006/01/02 - 15:04:05"), p.StatusCode, p.Latency, p.ClientIP, p.Method, path)
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
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
