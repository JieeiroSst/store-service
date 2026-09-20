package port

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
)

type ValidateCouponInput struct {
	Code           string
	UserID         int64
	PurchaseAmount float64
	CategoryIDs    []int64
	ProductIDs     []int64
}

type ApplyCouponInput struct {
	ValidateCouponInput
	OrderID int64
}

type CouponUsecase interface {
	CreateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error)
	GetCoupon(ctx context.Context, id int64) (*model.Coupon, error)
	GetCouponByCode(ctx context.Context, code string) (*model.Coupon, error)
	ListCoupons(ctx context.Context) ([]model.Coupon, error)
	UpdateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error)
	DeleteCoupon(ctx context.Context, id int64) error
	ValidateCoupon(ctx context.Context, in ValidateCouponInput) (*model.Coupon, float64, error)
	ApplyCoupon(ctx context.Context, in ApplyCouponInput) (*model.CouponUsage, error)
	DeactivateStaleCoupons(ctx context.Context) (int64, error)
}

type CouponRestrictionUsecase interface {
	CreateRestriction(ctx context.Context, restriction *model.CouponRestriction) (*model.CouponRestriction, error)
	GetRestriction(ctx context.Context, id int64) (*model.CouponRestriction, error)
	ListRestrictionsByCoupon(ctx context.Context, couponID int64) ([]model.CouponRestriction, error)
	UpdateRestriction(ctx context.Context, restriction *model.CouponRestriction) (*model.CouponRestriction, error)
	DeleteRestriction(ctx context.Context, id int64) error
}

type CouponUsageUsecase interface {
	ListUsagesByCoupon(ctx context.Context, couponID int64) ([]model.CouponUsage, error)
	ListUsagesByUser(ctx context.Context, userID int64) ([]model.CouponUsage, error)
}

type UserCouponUsecase interface {
	AssignCoupon(ctx context.Context, userID, couponID int64) (*model.UserCoupon, error)
	ListByUser(ctx context.Context, userID int64) ([]model.UserCoupon, error)
	ListByCoupon(ctx context.Context, couponID int64) ([]model.UserCoupon, error)
	UseCoupon(ctx context.Context, id int64) (*model.UserCoupon, error)
	UnuseCoupon(ctx context.Context, id int64) (*model.UserCoupon, error)
	DeleteUserCoupon(ctx context.Context, id int64) error
}
