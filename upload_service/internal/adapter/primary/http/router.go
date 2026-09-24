package http

import (
	"net/http"
	"time"

	"context"
	"github.com/JIeeiroSst/upload-service/config"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In

	Config  *config.Config
	Auth    *Authenticator
	Handler *Handler
	Mongo   *mongo.Client
}

func NewRouter(p RouterParams) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery(), cors(p.Config.Server.CORSOrigins))

	engine.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := p.Mongo.Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	up := engine.Group("/api/v1/upload", p.Auth.Middleware())
	up.POST("", p.Handler.Create)
	up.GET("", p.Handler.List)
	up.GET("/:id", p.Handler.Get)
	up.GET("/:id/content", p.Handler.Content)
	up.PUT("/:id", p.Handler.Replace)
	up.DELETE("/:id", p.Handler.Delete)
	return engine
}

func cors(origins []string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, o := range origins {
		allowed[o] = true
	}
	return func(c *gin.Context) {
		if o := c.GetHeader("Origin"); o != "" && allowed[o] {
			c.Header("Access-Control-Allow-Origin", o)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}
		c.Next()
	}
}

var Module = fx.Options(fx.Provide(NewAuthenticator, NewHandler, NewRouter))
