package http

import (
	"net/http"

	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListInventory(c *gin.Context) {
	page, err := pageRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	items, err := h.inventory.ListByMachine(c.Request.Context(), c.Param("id"), page)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPage(items, page, toInventory))
}

func (h *Handler) AssignSlot(c *gin.Context) {
	var req slotRequest
	if !bind(c, &req) {
		return
	}
	inv, err := h.inventory.AssignSlot(c.Request.Context(), c.Param("id"), c.Param("slot"), port.SlotInput{
		ProductID:    req.ProductID,
		Quantity:     req.Quantity,
		MaxCapacity:  req.MaxCapacity,
		LowThreshold: req.LowThreshold,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInventory(inv))
}

func (h *Handler) ListLowInventory(c *gin.Context) {
	page, err := pageRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	items, err := h.inventory.ListLow(c.Request.Context(), c.Query("machine_id"), page)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPage(items, page, toInventory))
}

func (h *Handler) Restock(c *gin.Context) {
	var req restockRequest
	if !bind(c, &req) {
		return
	}
	inv, err := h.inventory.Restock(c.Request.Context(), c.Param("id"), req.Quantity)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInventory(inv))
}
