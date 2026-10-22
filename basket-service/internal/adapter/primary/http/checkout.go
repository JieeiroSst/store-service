package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListBasketLinesOfBasket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if !h.authorizeBasket(c, id) {
		return
	}
	lines, err := h.basketLine.ListBasketLinesByBasket(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, lines)
}

func (h *Handler) CheckoutBasket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		WalletID          string `json:"wallet_id"`
		Pin               string `json:"pin"`
		ShippingAddressID int    `json:"shipping_address_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.authorizeBasket(c, id) {
		return
	}
	basket, err := h.basket.GetBasket(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	result, err := h.checkout.Checkout(c.Request.Context(), port.CheckoutCommand{
		BasketID: id, UserID: basket.UserID, WalletID: req.WalletID, Pin: req.Pin, ShippingAddressID: req.ShippingAddressID,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
