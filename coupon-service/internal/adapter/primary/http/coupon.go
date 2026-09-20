package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type couponRequest struct {
	Code              string    `json:"code" binding:"required"`
	Type              string    `json:"type" binding:"required"`
	DiscountValue     float64   `json:"discount_value" binding:"required"`
	MinimumPurchase   float64   `json:"minimum_purchase"`
	MaxDiscountAmount *float64  `json:"max_discount_amount"`
	Description       string    `json:"description"`
	StartDate         time.Time `json:"start_date" binding:"required"`
	EndDate           time.Time `json:"end_date" binding:"required"`
	IsActive          bool      `json:"is_active"`
	MaxUses           *int      `json:"max_uses"`
}

func (r couponRequest) toModel() *model.Coupon {
	return &model.Coupon{
		Code:              r.Code,
		Type:              model.CouponType(r.Type),
		DiscountValue:     r.DiscountValue,
		MinimumPurchase:   r.MinimumPurchase,
		MaxDiscountAmount: r.MaxDiscountAmount,
		Description:       r.Description,
		StartDate:         r.StartDate,
		EndDate:           r.EndDate,
		IsActive:          r.IsActive,
		MaxUses:           r.MaxUses,
	}
}

func (h *Handler) CreateCoupon(c *gin.Context) {
	var req couponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, err := h.coupons.CreateCoupon(c.Request.Context(), req.toModel())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, coupon)
}

func (h *Handler) ListCoupons(c *gin.Context) {
	coupons, err := h.coupons.ListCoupons(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, coupons)
}

func (h *Handler) GetCoupon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	coupon, err := h.coupons.GetCoupon(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, coupon)
}

func (h *Handler) GetCouponByCode(c *gin.Context) {
	coupon, err := h.coupons.GetCouponByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, coupon)
}

func (h *Handler) UpdateCoupon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req couponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon := req.toModel()
	coupon.ID = id
	result, err := h.coupons.UpdateCoupon(c.Request.Context(), coupon)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteCoupon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.coupons.DeleteCoupon(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type validateCouponRequest struct {
	Code           string  `json:"code" binding:"required"`
	UserID         int64   `json:"user_id"`
	PurchaseAmount float64 `json:"purchase_amount"`
	CategoryIDs    []int64 `json:"category_ids"`
	ProductIDs     []int64 `json:"product_ids"`
}

func (r validateCouponRequest) toInput() port.ValidateCouponInput {
	return port.ValidateCouponInput{
		Code:           r.Code,
		UserID:         r.UserID,
		PurchaseAmount: r.PurchaseAmount,
		CategoryIDs:    r.CategoryIDs,
		ProductIDs:     r.ProductIDs,
	}
}

type validateCouponResponse struct {
	Coupon   *model.Coupon `json:"coupon"`
	Discount float64       `json:"discount"`
}

func (h *Handler) ValidateCoupon(c *gin.Context) {
	var req validateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, discount, err := h.coupons.ValidateCoupon(c.Request.Context(), req.toInput())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, validateCouponResponse{Coupon: coupon, Discount: discount})
}

type applyCouponRequest struct {
	validateCouponRequest
	OrderID int64 `json:"order_id" binding:"required"`
}

func (h *Handler) ApplyCoupon(c *gin.Context) {
	var req applyCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usage, err := h.coupons.ApplyCoupon(c.Request.Context(), port.ApplyCouponInput{
		ValidateCouponInput: req.toInput(),
		OrderID:             req.OrderID,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, usage)
}
