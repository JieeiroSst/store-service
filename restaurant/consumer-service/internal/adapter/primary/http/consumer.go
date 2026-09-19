package http

import (
	"net/http"

	"github.com/JIeeiroSst/consumer-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) FindConsumer(c *gin.Context) {
	var pagination logger.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	consumers, err := h.consumer.Find(c, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": consumers})
}

func (h *Handler) ConsumerOrder(c *gin.Context) {
	var order model.PlaceOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.consumer.PlaceOrder(c, order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
