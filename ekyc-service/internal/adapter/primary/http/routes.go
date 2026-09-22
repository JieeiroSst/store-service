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
		ekyc := api.Group("/ekyc")
		{
			ekyc.POST("/:user_id/citizen-card", h.SubmitCitizenCard)
			ekyc.POST("/:user_id/nfc-chip", h.SubmitNFCChip)
			ekyc.POST("/:user_id/face-scan", h.SubmitFaceScan)
			ekyc.POST("/:user_id/verify", h.Verify)
			ekyc.GET("/:user_id", h.GetStatus)
		}
	}

	return engine
}
