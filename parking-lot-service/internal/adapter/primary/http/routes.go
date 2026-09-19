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
		api.GET("/spots/available", h.AvailableSpots)
		api.GET("/rates", h.ListRates)

		api.POST("/check-in", h.CheckIn)
		api.POST("/check-out", h.CheckOut)

		api.GET("/tickets/:id", h.GetTicket)
		api.GET("/history", h.ListHistory)
	}

	return engine
}
