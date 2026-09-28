package http

import (
	"net/http"

	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type ghnWebhookRequest struct {
	OrderCode       string `json:"OrderCode"`
	ClientOrderCode string `json:"ClientOrderCode"`
	Type            string `json:"Type"`
	Status          string `json:"Status"`
	Reason          string `json:"Reason"`
	Time            string `json:"Time"`
}

func (h *Handler) GHNWebhook(c *gin.Context) {
	var req ghnWebhookRequest
	if !bind(c, &req) {
		return
	}
	if err := h.webhooks.HandleGHN(c.Request.Context(), c.Query("token"), port.CarrierWebhook(req)); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
