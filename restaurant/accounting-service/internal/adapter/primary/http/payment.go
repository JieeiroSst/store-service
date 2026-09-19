package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPayment(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("orderID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.payment.Get(c, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payment})
}

// MarkPaid is the "customer settled the bill" action a staff member takes
// at checkout.
func (h *Handler) MarkPaid(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("orderID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.payment.MarkPaid(c, orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
