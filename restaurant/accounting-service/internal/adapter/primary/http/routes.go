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
	accounting := api.Group("/accounting")
	{
		accounting.POST("", h.AuthCart)
	}

	payments := api.Group("/payments")
	{
		payments.GET("/:orderID", h.GetPayment)
		payments.PATCH("/:orderID/pay", h.MarkPaid)
	}
}
