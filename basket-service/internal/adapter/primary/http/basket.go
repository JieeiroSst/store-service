package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateBasket(c *gin.Context) {
	var basket model.Basket
	if err := c.ShouldBindJSON(&basket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	owner, ok := callerID(c)
	if !ok {
		return
	}
	basket.UserID = owner // a basket always belongs to the caller
	result, err := h.basket.CreateBasket(c.Request.Context(), &basket)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetBasket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.basket.GetBasket(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	if !isOwner(c, result.UserID) {
		writeError(c, common.ErrNotFound)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListBaskets(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		owner, ok := callerID(c)
		if !ok {
			return
		}
		result, err := h.basket.ListBasketsByUser(c.Request.Context(), owner)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}
	result, err := h.basket.ListBaskets(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UpdateBasket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var basket model.Basket
	if err := c.ShouldBindJSON(&basket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	basket.ID = id
	existing, err := h.basket.GetBasket(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	if !isOwner(c, existing.UserID) {
		writeError(c, common.ErrNotFound)
		return
	}
	basket.UserID = existing.UserID // ownership can't be transferred
	result, err := h.basket.UpdateBasket(c.Request.Context(), &basket)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteBasket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if !h.authorizeBasket(c, id) {
		return
	}
	if err := h.basket.DeleteBasket(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
