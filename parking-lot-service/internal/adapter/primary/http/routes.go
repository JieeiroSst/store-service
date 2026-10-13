package http

import (
	"github.com/JIeeiroSst/parking-lot-service/pkg/authn"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	api := engine.Group("/api/v1", authenticate(authn.FromEnv()))
	{
		api.GET("/spots/available", h.AvailableSpots)
		api.GET("/rates", h.ListRates)

		staff := api.Group("", requireAdmin())
		staff.POST("/check-in", h.CheckIn)
		staff.POST("/check-out", h.CheckOut)
		staff.GET("/tickets/:id", h.GetTicket)
		staff.GET("/history", h.ListHistory)
	}

	return engine
}
