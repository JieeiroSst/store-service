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
		ekyc := api.Group("/ekyc/:user_id", h.Authorize)
		{
			ekyc.POST("/citizen-card", h.SubmitCitizenCard)
			ekyc.POST("/nfc-chip", h.SubmitNFCChip)
			ekyc.POST("/face-scan", h.SubmitFaceScan)
			ekyc.POST("/verify", h.Verify)
			ekyc.GET("", h.GetStatus)
		}
	}

	return engine
}
