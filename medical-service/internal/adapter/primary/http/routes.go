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
		medicines := api.Group("/medicines")
		{
			medicines.POST("", h.CreateMedicine)
			medicines.GET("", h.ListMedicines)
			medicines.GET("/:id", h.GetMedicine)
			medicines.GET("/:id/stock", h.GetStock)
			medicines.GET("/:id/movements", h.ListMovements)
			medicines.POST("/:id/batches", h.ReceiveBatch)
		}

		api.POST("/batches/:id/adjust", h.AdjustBatch)

		inventory := api.Group("/inventory")
		{
			inventory.GET("/expiring", h.ListExpiring)
			inventory.GET("/low-stock", h.ListLowStock)
			inventory.POST("/dispose-expired", h.DisposeExpired)
		}

		interactions := api.Group("/interactions")
		{
			interactions.POST("", h.CreateInteraction)
			interactions.GET("", h.ListInteractions)
			interactions.POST("/check", h.CheckSafety)
		}

		terminology := api.Group("/terminology")
		{
			terminology.GET("", h.ListTerminology)
			terminology.POST("/aliases", h.AddAlias)
			terminology.POST("/allergen-groups", h.AddAllergenGroup)
		}

		dispenses := api.Group("/dispenses")
		{
			dispenses.POST("", h.CreateDispense)
			dispenses.GET("", h.ListDispenses)
			dispenses.GET("/:id", h.GetDispense)
		}
	}

	return engine
}
