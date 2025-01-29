package http

import (
	"context"

	pd "github.com/JIeeiroSst/coupon-service/gateway/proto"
	"github.com/JIeeiroSst/coupon-service/internal/usecase"
)

type Handler struct {
	usecase *usecase.Usecase
	pd.UnimplementedCouponServiceServer
}

func NewHandler(usecase *usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) CreateCoupon(ctx context.Context, in *pd.CreateCouponRequest) (*pd.Coupon, error) {
	dtoCoupon := buildDtoCoupon(in)
	coupon, err := h.usecase.CreateCoupon(ctx, &dtoCoupon)
	if err != nil {
		return nil, err
	}
	pbCoupon := buildPbCoupon(coupon)
	return &pbCoupon, nil
}

func (h *Handler) GetCoupon(ctx context.Context, in *pd.GetCouponRequest) (*pd.Coupon, error) {
	return nil, nil
}

func (h *Handler) UpdateCoupon(ctx context.Context, in *pd.UpdateCouponRequest) (*pd.Coupon, error) {
	return nil, nil
}

func (h *Handler) DeleteCoupon(ctx context.Context, in *pd.DeleteCouponRequest) (*pd.DeleteCouponResponse, error) {
	return nil, nil
}

func (h *Handler) ListCoupons(ctx context.Context, in *pd.ListCouponsRequest) (*pd.ListCouponsResponse, error) {
	return nil, nil
}

func (h *Handler) CreateCouponRestriction(ctx context.Context, in *pd.CreateCouponRestrictionRequest) (*pd.CouponRestriction, error) {
	return nil, nil
}

func (h *Handler) GetCouponRestriction(ctx context.Context, in *pd.GetCouponRestrictionRequest) (*pd.CouponRestriction, error) {
	return nil, nil
}

func (h *Handler) UpdateCouponRestriction(ctx context.Context, in *pd.UpdateCouponRestrictionRequest) (*pd.CouponRestriction, error) {
	return nil, nil
}

func (h *Handler) DeleteCouponRestriction(ctx context.Context, in *pd.DeleteCouponRestrictionRequest) (*pd.DeleteCouponRestrictionResponse, error) {
	return nil, nil
}

func (h *Handler) ListCouponRestrictions(ctx context.Context, in *pd.ListCouponRestrictionsRequest) (*pd.ListCouponRestrictionsResponse, error) {
	return nil, nil
}

func (h *Handler) CreateUserCoupon(ctx context.Context, in *pd.CreateUserCouponRequest) (*pd.UserCoupon, error) {
	return nil, nil
}

func (h *Handler) GetUserCoupon(ctx context.Context, in *pd.GetUserCouponRequest) (*pd.UserCoupon, error) {
	return nil, nil
}

func (h *Handler) ListUserCoupons(ctx context.Context, in *pd.ListUserCouponsRequest) (*pd.ListUserCouponsResponse, error) {
	return nil, nil
}

func (h *Handler) UseUserCoupon(ctx context.Context, in *pd.UseUserCouponRequest) (*pd.UserCoupon, error) {
	return nil, nil
}

func (h *Handler) UnuseUserCoupon(ctx context.Context, in *pd.UnuseUserCouponRequest) (*pd.UserCoupon, error) {
	return nil, nil
}

func (h *Handler) ListCouponUsages(ctx context.Context, in *pd.ListCouponUsagesRequest) (*pd.ListCouponUsagesResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsages(ctx context.Context, in *pd.ListUserCouponUsagesRequest) (*pd.ListUserCouponUsagesResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrder(ctx context.Context, in *pd.ListUserCouponUsagesByOrderRequest) (*pd.ListUserCouponUsagesByOrderResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByCoupon(ctx context.Context, in *pd.ListUserCouponUsagesByCouponRequest) (*pd.ListUserCouponUsagesByCouponResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrderAndCoupon(ctx context.Context, in *pd.ListUserCouponUsagesByOrderAndCouponRequest) (*pd.ListUserCouponUsagesByOrderAndCouponResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrderAndCouponAndUser(ctx context.Context, in *pd.ListUserCouponUsagesByOrderAndCouponAndUserRequest) (*pd.ListUserCouponUsagesByOrderAndCouponAndUserResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrderAndUser(ctx context.Context, in *pd.ListUserCouponUsagesByOrderAndUserRequest) (*pd.ListUserCouponUsagesByOrderAndUserResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByCouponAndUser(ctx context.Context, in *pd.ListUserCouponUsagesByCouponAndUserRequest) (*pd.ListUserCouponUsagesByCouponAndUserResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByUser(ctx context.Context, in *pd.ListUserCouponUsagesByUserRequest) (*pd.ListUserCouponUsagesByUserResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrderAndUserAndCoupon(ctx context.Context, in *pd.ListUserCouponUsagesByOrderAndUserAndCouponRequest) (*pd.ListUserCouponUsagesByOrderAndUserAndCouponResponse, error) {
	return nil, nil
}
