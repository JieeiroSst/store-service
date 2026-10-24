package http

import (
	"net/http"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) AuthCart(c *gin.Context) {
	var cart model.AuthCart
	if err := c.ShouldBindJSON(&cart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authCart.PlaceOrder(c, cart); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
