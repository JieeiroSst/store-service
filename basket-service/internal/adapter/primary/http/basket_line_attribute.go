package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateBasketLineAttribute(c *gin.Context) {
	var attribute model.BasketLineAttribute
	if err := c.ShouldBindJSON(&attribute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.basketLineAttr.CreateBasketLineAttribute(c.Request.Context(), &attribute)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetBasketLineAttribute(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.basketLineAttr.GetBasketLineAttribute(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListBasketLineAttributes(c *gin.Context) {
	result, err := h.basketLineAttr.ListBasketLineAttributes(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UpdateBasketLineAttribute(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var attribute model.BasketLineAttribute
	if err := c.ShouldBindJSON(&attribute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	attribute.ID = id
	result, err := h.basketLineAttr.UpdateBasketLineAttribute(c.Request.Context(), &attribute)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteBasketLineAttribute(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.basketLineAttr.DeleteBasketLineAttribute(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
