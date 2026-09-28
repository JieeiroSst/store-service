package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/adapter/primary/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.order.GetOrder(c.Request.Context(), id)
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

func (h *Handler) ListOrders(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		owner, ok := callerID(c)
		if !ok {
			return
		}
		result, err := h.order.ListOrdersByUser(c.Request.Context(), owner)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}
	result, err := h.order.ListOrders(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
