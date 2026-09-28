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
	engine.POST("/api/v1/webhooks/ghn", h.GHNWebhook)

	api := engine.Group("/api/v1", h.authenticate)
	{
		shipments := api.Group("/shipments")
		{
			shipments.POST("/quote", h.Quote)
			shipments.POST("", h.CreateShipment)
			shipments.GET("", h.ListShipments)
			shipments.GET("/:id", h.GetShipment)
			shipments.GET("/by-client-code/:code", h.GetShipmentByClientOrderCode)
			shipments.POST("/:id/place", h.PlaceShipment)
			shipments.POST("/:id/cancel", h.CancelShipment)
			shipments.POST("/:id/sync", h.SyncShipment)
		}

		warehouses := api.Group("/warehouses")
		{
			warehouses.POST("", h.CreateWarehouse)
			warehouses.GET("", h.ListWarehouses)
			warehouses.PATCH("/:id", h.SetWarehouseActive)
		}

		locations := api.Group("/locations")
		{
			locations.GET("/provinces", h.Provinces)
			locations.GET("/districts", h.Districts)
			locations.GET("/wards", h.Wards)
		}
	}

	return engine
}
