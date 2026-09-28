package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetOrder(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	order, err := h.orders.GetOrder(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	list, err := h.orders.ListOrders(c.Request.Context(), parseList(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list.Items, "total": list.Total})
}

func (h *Handler) SyncOrders(c *gin.Context) {
	result, err := h.orders.SyncOrders(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
