package http

import (
	"net/http"

	"github.com/JIeeiroSst/delivery-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterDelivery(c *gin.Context) {
	var delivery model.Delivery
	if err := c.ShouldBind(&delivery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.delivery.Create(c, &delivery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": delivery})
}

func (h *Handler) All(c *gin.Context) {
	var pagination logger.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deliveries, err := h.delivery.FindAll(c, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": deliveries})
}
