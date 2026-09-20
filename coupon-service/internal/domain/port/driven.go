package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
)

type CouponRepository interface {
	Create(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error)
	GetByID(ctx context.Context, id int64) (*model.Coupon, error)
	GetByCode(ctx context.Context, code string) (*model.Coupon, error)
	List(ctx context.Context) ([]model.Coupon, error)
	Update(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error)
	Delete(ctx context.Context, id int64) error
	IncrementUsage(ctx context.Context, id int64) error
	DeactivateStale(ctx context.Context, now time.Time) (int64, error)
}

type CouponRestrictionRepository interface {
	Create(ctx context.Context, restriction *model.CouponRestriction) (*model.CouponRestriction, error)
	GetByID(ctx context.Context, id int64) (*model.CouponRestriction, error)
	ListByCoupon(ctx context.Context, couponID int64) ([]model.CouponRestriction, error)
	Update(ctx context.Context, restriction *model.CouponRestriction) (*model.CouponRestriction, error)
	Delete(ctx context.Context, id int64) error
}

type CouponUsageRepository interface {
	Create(ctx context.Context, usage *model.CouponUsage) error
	GetByOrderAndCoupon(ctx context.Context, orderID, couponID int64) (*model.CouponUsage, error)
	ListByCoupon(ctx context.Context, couponID int64) ([]model.CouponUsage, error)
	ListByUser(ctx context.Context, userID int64) ([]model.CouponUsage, error)
}

type UserCouponRepository interface {
	Create(ctx context.Context, userCoupon *model.UserCoupon) (*model.UserCoupon, error)
	GetByID(ctx context.Context, id int64) (*model.UserCoupon, error)
	GetByUserAndCoupon(ctx context.Context, userID, couponID int64) (*model.UserCoupon, error)
	Update(ctx context.Context, userCoupon *model.UserCoupon) (*model.UserCoupon, error)
	Delete(ctx context.Context, id int64) error
	ListByUser(ctx context.Context, userID int64) ([]model.UserCoupon, error)
	ListByCoupon(ctx context.Context, couponID int64) ([]model.UserCoupon, error)
}
