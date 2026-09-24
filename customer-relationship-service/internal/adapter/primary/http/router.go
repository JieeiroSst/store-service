package http

import (
	"net/http"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In

	Config    *config.Config
	Auth      *Authenticator
	Resources []Resource `group:"resources"`
}

func NewRouter(p RouterParams) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), RequestLogMiddleware(), CORSMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := engine.Group("/api/v1")
	api.Use(p.Auth.Middleware())
	for _, r := range p.Resources {
		r.Register(api)
	}
	return engine
}
