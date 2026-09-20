package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListUsagesByCoupon(c *gin.Context) {
	couponID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon id"})
		return
	}
	usages, err := h.usages.ListUsagesByCoupon(c.Request.Context(), couponID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, usages)
}

func (h *Handler) ListUsagesByUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	usages, err := h.usages.ListUsagesByUser(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, usages)
}
