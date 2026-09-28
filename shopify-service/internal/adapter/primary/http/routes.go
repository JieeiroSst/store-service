package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	api := engine.Group("/api/v1")
	{
		products := api.Group("/products")
		{
			products.POST("", h.CreateProduct)
			products.GET("", h.ListProducts)
			products.GET("/:id", h.GetProduct)
			products.POST("/sync", h.SyncProducts)
		}

		orders := api.Group("/orders")
		{
			orders.GET("", h.ListOrders)
			orders.GET("/:id", h.GetOrder)
			orders.POST("/sync", h.SyncOrders)
		}

		api.POST("/webhooks", h.HandleWebhook)
	}

	return engine
}
