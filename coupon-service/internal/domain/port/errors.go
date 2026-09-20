package port

import "errors"

var (
	ErrNotFound               = errors.New("resource not found")
	ErrInvalidCouponType      = errors.New("invalid coupon type")
	ErrInvalidRestrictionType = errors.New("invalid restriction type")
	ErrInvalidDateRange       = errors.New("end date must be after start date")

	ErrCouponInactive          = errors.New("coupon is inactive")
	ErrCouponExpired           = errors.New("coupon is not valid at this time")
	ErrMinimumPurchaseNotMet   = errors.New("purchase amount is below the coupon's minimum purchase")
	ErrUsageLimitReached       = errors.New("coupon has reached its usage limit")
	ErrCouponRestricted        = errors.New("coupon does not apply to this category/product")
	ErrCouponNotAssignedToUser = errors.New("coupon is not assigned to this user")
	ErrCouponAlreadyUsedByUser = errors.New("user has already used this coupon")
	ErrOrderAlreadyUsedCoupon  = errors.New("this order has already used this coupon")
)
