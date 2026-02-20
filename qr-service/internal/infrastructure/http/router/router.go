package router

import (
	"github.com/JIeeiroSst/qr-service/internal/infrastructure/http/handler"
	"github.com/JIeeiroSst/qr-service/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter(qrHandler *handler.QRHandler) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/qr/scan/:shortCode", qrHandler.Redirect)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		qr := v1.Group("/qr")
		{
			qr.POST("", qrHandler.Generate)
			qr.GET("", qrHandler.List)
			qr.GET("/:id", qrHandler.GetByID)
			qr.PUT("/:id", qrHandler.Update)
			qr.PATCH("/:id/content", qrHandler.UpdateContent)
			qr.DELETE("/:id", qrHandler.Delete)
			qr.POST("/:id/regenerate", qrHandler.Regenerate)
			qr.GET("/:id/history", qrHandler.GetHistory)
			qr.GET("/:id/stats", qrHandler.GetStats)
		}
	}

	return r
}
