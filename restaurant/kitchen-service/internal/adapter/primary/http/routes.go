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
	category := api.Group("/category")
	{
		category.GET("", h.FindCategory)
		category.POST("", h.CreateCategory)
	}

	food := api.Group("/food")
	{
		food.GET("", h.FindFood)
		food.POST("", h.CreateFood)
		food.GET("/batch", h.FindFoodByIDs)
	}

	kitchen := api.Group("/kitchen")
	{
		kitchen.GET("", h.FindKitchen)
		kitchen.PATCH("/:id/status", h.UpdateKitchenStatus)
	}
}
