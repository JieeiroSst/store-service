package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	coupons      port.CouponUsecase
	restrictions port.CouponRestrictionUsecase
	usages       port.CouponUsageUsecase
	userCoupons  port.UserCouponUsecase
}

func NewHandler(
	coupons port.CouponUsecase,
	restrictions port.CouponRestrictionUsecase,
	usages port.CouponUsageUsecase,
	userCoupons port.UserCouponUsecase,
) *Handler {
	return &Handler{coupons: coupons, restrictions: restrictions, usages: usages, userCoupons: userCoupons}
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrCouponInactive),
		errors.Is(err, port.ErrCouponExpired),
		errors.Is(err, port.ErrMinimumPurchaseNotMet),
		errors.Is(err, port.ErrUsageLimitReached),
		errors.Is(err, port.ErrCouponRestricted),
		errors.Is(err, port.ErrCouponNotAssignedToUser),
		errors.Is(err, port.ErrCouponAlreadyUsedByUser),
		errors.Is(err, port.ErrOrderAlreadyUsedCoupon):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrInvalidCouponType),
		errors.Is(err, port.ErrInvalidRestrictionType),
		errors.Is(err, port.ErrInvalidDateRange):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
