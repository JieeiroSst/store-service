package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type assignCouponRequest struct {
	UserID   int64 `json:"user_id" binding:"required"`
	CouponID int64 `json:"coupon_id" binding:"required"`
}

func (h *Handler) AssignCoupon(c *gin.Context) {
	var req assignCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCoupon, err := h.userCoupons.AssignCoupon(c.Request.Context(), req.UserID, req.CouponID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, userCoupon)
}

func (h *Handler) ListUserCouponsByUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	userCoupons, err := h.userCoupons.ListByUser(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, userCoupons)
}

func (h *Handler) ListUserCouponsByCoupon(c *gin.Context) {
	couponID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon id"})
		return
	}
	userCoupons, err := h.userCoupons.ListByCoupon(c.Request.Context(), couponID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, userCoupons)
}

func (h *Handler) UseCoupon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userCoupon, err := h.userCoupons.UseCoupon(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, userCoupon)
}

func (h *Handler) UnuseCoupon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userCoupon, err := h.userCoupons.UnuseCoupon(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, userCoupon)
}

func (h *Handler) DeleteUserCoupon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.userCoupons.DeleteUserCoupon(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
