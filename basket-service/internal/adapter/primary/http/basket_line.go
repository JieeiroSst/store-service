package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateBasketLine(c *gin.Context) {
	var line model.BasketLine
	if err := c.ShouldBindJSON(&line); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.basketLine.CreateBasketLine(c.Request.Context(), &line)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetBasketLine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.basketLine.GetBasketLine(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListBasketLines(c *gin.Context) {
	result, err := h.basketLine.ListBasketLines(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UpdateBasketLine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var line model.BasketLine
	if err := c.ShouldBindJSON(&line); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	line.ID = id
	result, err := h.basketLine.UpdateBasketLine(c.Request.Context(), &line)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteBasketLine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.basketLine.DeleteBasketLine(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
