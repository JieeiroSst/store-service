package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/kitchen-service/common"
	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) FindKitchen(c *gin.Context) {
	var pagination logger.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kitchens, err := h.kitchen.Find(c, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": kitchens})
}

type updatePrepStatusRequest struct {
	Status string `json:"status" form:"status"`
}

// UpdateKitchenStatus moves a kitchen ticket through the prep workflow:
// received -> preparing -> ready -> served.
func (h *Handler) UpdateKitchenStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req updatePrepStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.kitchen.UpdateStatus(c, id, req.Status); err != nil {
		if errors.Is(err, common.ErrInvalidPrepStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
