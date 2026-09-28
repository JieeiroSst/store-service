package http

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxWebhookBody = 5 << 20

func (h *Handler) HandleWebhook(c *gin.Context) {
	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBody))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}
	if err := h.webhooks.HandleWebhook(c.Request.Context(), payload, c.Request.Header); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
