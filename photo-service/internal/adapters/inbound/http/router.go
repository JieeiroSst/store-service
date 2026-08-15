package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(handler *CompositionHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/v1")
	{
		v1.POST("/compositions", handler.Create)
		v1.GET("/compositions/:id", handler.Get)
	}

	return r
}
