package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.Default()
	engine.Use(CORSMiddleware())
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api/v1")
	RegisterRoutes(api, h)

	return engine
}

func RegisterRoutes(api *gin.RouterGroup, h *Handler) {
	consumer := api.Group("/consumer")
	{
		consumer.GET("", h.FindConsumer)
		consumer.POST("", h.ConsumerOrder)
	}
}
