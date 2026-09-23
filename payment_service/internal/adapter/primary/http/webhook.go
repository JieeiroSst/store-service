package http

import (
	"io"
	"net/http"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) HandleWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	provider := model.Provider(c.Param("provider"))
	if err := h.payments.HandleWebhook(c.Request.Context(), provider, payload, c.Request.Header); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"received": true})
}
