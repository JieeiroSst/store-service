package http

import (
	"net/http"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) FindCategory(c *gin.Context) {
	categories, err := h.category.Find(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var category model.Category
	if err := c.ShouldBind(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.category.Create(c, &category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
