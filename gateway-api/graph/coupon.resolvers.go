package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) CouponCreateCoupon(ctx context.Context, input model.CouponCouponInput) (*model.CouponCoupon, error) {
	return restCall[*model.CouponCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "POST", Path: "/api/v1/coupons", Body: input}, "")
}

func (r *mutationResolver) CouponUpdateCoupon(ctx context.Context, id int, input model.CouponCouponInput) (*model.CouponCoupon, error) {
	return restCall[*model.CouponCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/coupons/%v", id), Body: input}, "")
}

func (r *mutationResolver) CouponDeleteCoupon(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "coupon-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/coupons/%v", id)}, "")
}

func (r *mutationResolver) CouponCreateRestriction(ctx context.Context, id int, input model.CouponRestrictionInput) (*model.CouponRestriction, error) {
	return restCall[*model.CouponRestriction](ctx, r.Proxy, "coupon-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/coupons/%v/restrictions", id), Body: input}, "")
}

func (r *mutationResolver) CouponUpdateRestriction(ctx context.Context, id int, input model.CouponRestrictionInput) (*model.CouponRestriction, error) {
	return restCall[*model.CouponRestriction](ctx, r.Proxy, "coupon-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/restrictions/%v", id), Body: input}, "")
}

func (r *mutationResolver) CouponDeleteRestriction(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "coupon-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/restrictions/%v", id)}, "success")
}

func (r *mutationResolver) ApplyCoupon(ctx context.Context, input model.CouponApplyInput) (*model.CouponUsage, error) {
	return restCall[*model.CouponUsage](ctx, r.Proxy, "coupon-service", rest.Request{Method: "POST", Path: "/api/v1/coupons/apply", Body: withField(input, "user_id", userIDInt(ctx))}, "")
}

func (r *mutationResolver) CouponAssign(ctx context.Context, input model.CouponAssignInput) (*model.CouponUserCoupon, error) {
	return restCall[*model.CouponUserCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "POST", Path: "/api/v1/user-coupons", Body: input}, "")
}

func (r *mutationResolver) CouponUseUserCoupon(ctx context.Context, id int) (*model.CouponUserCoupon, error) {
	return restCall[*model.CouponUserCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/user-coupons/%v/use", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CouponUnuseUserCoupon(ctx context.Context, id int) (*model.CouponUserCoupon, error) {
	return restCall[*model.CouponUserCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/user-coupons/%v/unuse", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CouponDeleteUserCoupon(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "coupon-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/user-coupons/%v", id)}, "success")
}

func (r *queryResolver) CouponCoupons(ctx context.Context) ([]*model.CouponCoupon, error) {
	return restCall[[]*model.CouponCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: "/api/v1/coupons"}, "")
}

func (r *queryResolver) CouponCoupon(ctx context.Context, id int) (*model.CouponCoupon, error) {
	return restCall[*model.CouponCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/coupons/%v", id)}, "")
}

func (r *queryResolver) CouponByCode(ctx context.Context, code string) (*model.CouponCoupon, error) {
	return restCall[*model.CouponCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/coupons/code/%v", code)}, "")
}

func (r *queryResolver) CouponRestrictions(ctx context.Context, id int) ([]*model.CouponRestriction, error) {
	return restCall[[]*model.CouponRestriction](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/coupons/%v/restrictions", id)}, "")
}

func (r *queryResolver) CouponUsages(ctx context.Context, id int) ([]*model.CouponUsage, error) {
	return restCall[[]*model.CouponUsage](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/coupons/%v/usages", id)}, "")
}

func (r *queryResolver) CouponUserCouponsByCoupon(ctx context.Context, id int) ([]*model.CouponUserCoupon, error) {
	return restCall[[]*model.CouponUserCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/coupons/%v/user-coupons", id)}, "")
}

func (r *queryResolver) ValidateCoupon(ctx context.Context, input model.CouponValidateInput) (*model.CouponValidation, error) {
	return restCall[*model.CouponValidation](ctx, r.Proxy, "coupon-service", rest.Request{Method: "POST", Path: "/api/v1/coupons/validate", Body: withField(input, "user_id", userIDInt(ctx))}, "")
}

func (r *queryResolver) CouponUserCoupons(ctx context.Context, userID int) ([]*model.CouponUserCoupon, error) {
	return restCall[[]*model.CouponUserCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: "/api/v1/user-coupons", Query: restQuery("user_id", userID)}, "")
}

func (r *queryResolver) MyCoupons(ctx context.Context) ([]*model.CouponUserCoupon, error) {
	return restCall[[]*model.CouponUserCoupon](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: "/api/v1/user-coupons", Query: restQuery("user_id", userIDOf(ctx))}, "")
}

func (r *queryResolver) CouponUserUsages(ctx context.Context, userID int) ([]*model.CouponUsage, error) {
	return restCall[[]*model.CouponUsage](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/coupon-usages", userID)}, "")
}

func (r *queryResolver) MyCouponUsages(ctx context.Context) ([]*model.CouponUsage, error) {
	return restCall[[]*model.CouponUsage](ctx, r.Proxy, "coupon-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/coupon-usages", userIDOf(ctx))}, "")
}
