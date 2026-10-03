package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) CouponQuery() generated.CouponQueryResolver { return &couponQueryResolver{r} }

type couponQueryResolver struct{ *Resolver }

func (r *couponQueryResolver) Coupons(ctx context.Context, obj *model.CouponQuery) ([]*model.CouponCoupon, error) {
	return r.Clients.CouponService.Coupons(ctx)
}

func (r *couponQueryResolver) CouponByCode(ctx context.Context, obj *model.CouponQuery, code string) (*model.CouponCoupon, error) {
	return r.Clients.CouponService.CouponByCode(ctx, code)
}

func (r *couponQueryResolver) Coupon(ctx context.Context, obj *model.CouponQuery, id int) (*model.CouponCoupon, error) {
	return r.Clients.CouponService.Coupon(ctx, id)
}

func (r *couponQueryResolver) CouponRestrictions(ctx context.Context, obj *model.CouponQuery, id int) ([]*model.CouponCouponRestriction, error) {
	return r.Clients.CouponService.CouponRestrictions(ctx, id)
}

func (r *couponQueryResolver) CouponUsages(ctx context.Context, obj *model.CouponQuery, id int) ([]*model.CouponCouponUsage, error) {
	return r.Clients.CouponService.CouponUsages(ctx, id)
}

func (r *couponQueryResolver) CouponUserCoupons(ctx context.Context, obj *model.CouponQuery, id int) ([]*model.CouponUserCoupon, error) {
	return r.Clients.CouponService.CouponUserCoupons(ctx, id)
}

func (r *couponQueryResolver) UserCoupons(ctx context.Context, obj *model.CouponQuery, userID int) ([]*model.CouponUserCoupon, error) {
	return r.Clients.CouponService.UserCoupons(ctx, userID)
}

func (r *couponQueryResolver) UserCouponUsages(ctx context.Context, obj *model.CouponQuery, userID int) ([]*model.CouponCouponUsage, error) {
	return r.Clients.CouponService.UserCouponUsages(ctx, userID)
}
