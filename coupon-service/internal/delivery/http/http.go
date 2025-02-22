package http

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/dto"
	"github.com/JIeeiroSst/coupon-service/internal/usecase"
	couponServiceGrpc "github.com/JIeeiroSst/lib-gateway/coupon-service/gateway/coupon-service"

	"github.com/JIeeiroSst/utils/copy"
)

type Handler struct {
	usecase *usecase.Usecase
	couponServiceGrpc.UnimplementedCouponServiceServer
}

func NewHandler(usecase *usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) CreateCoupon(ctx context.Context, in *couponServiceGrpc.CreateCouponRequest) (*couponServiceGrpc.Coupon, error) {
	var coupon dto.Coupon
	if err := copy.CopyObject(&in, &coupon); err != nil {
		return nil, err
	}
	result, err := h.usecase.CouponUsecase.CreateCoupon(ctx, &coupon)
	if err != nil {
		return nil, err
	}
	var pbResult couponServiceGrpc.Coupon
	if err := copy.CopyObject(&result, &pbResult); err != nil {
		return nil, err
	}
	return &pbResult, nil
}

func (h *Handler) GetCoupon(ctx context.Context, in *couponServiceGrpc.GetCouponRequest) (*couponServiceGrpc.Coupon, error) {
	coupon, err := h.usecase.CouponUsecase.GetCoupons(ctx)
	if err != nil {
		return nil, err
	}
	var result couponServiceGrpc.Coupon
	if err := copy.CopyObject(&coupon, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *Handler) UpdateCoupon(ctx context.Context, in *couponServiceGrpc.UpdateCouponRequest) (*couponServiceGrpc.Coupon, error) {
	var coupon dto.Coupon
	if err := copy.CopyObject(&in, &coupon); err != nil {
		return nil, err
	}
	result, err := h.usecase.CouponUsecase.UpdateCoupon(ctx, coupon)
	if err != nil {
		return nil, err
	}

	var resultPd couponServiceGrpc.Coupon
	if err := copy.CopyObject(&result, &resultPd); err != nil {
		return nil, err
	}
	return &resultPd, nil
}

func (h *Handler) DeleteCoupon(ctx context.Context, in *couponServiceGrpc.DeleteCouponRequest) (*couponServiceGrpc.DeleteCouponResponse, error) {
	if in == nil {
		return &couponServiceGrpc.DeleteCouponResponse{
			Success: false,
		}, nil
	}
	if err := h.usecase.CouponUsecase.DeleteCoupon(ctx, int(in.Id)); err != nil {
		return &couponServiceGrpc.DeleteCouponResponse{
			Success: false,
		}, err
	}
	return &couponServiceGrpc.DeleteCouponResponse{
		Success: true,
	}, nil
}

func (h *Handler) ListCoupons(ctx context.Context, in *couponServiceGrpc.ListCouponsRequest) (*couponServiceGrpc.ListCouponsResponse, error) {
	coupon, err := h.usecase.CouponUsecase.GetCoupons(ctx)
	if err != nil {
		return nil, err
	}
	var result couponServiceGrpc.ListCouponsResponse
	if err := copy.CopyStructArrays(&coupon, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *Handler) CreateCouponRestriction(ctx context.Context, in *couponServiceGrpc.CreateCouponRestrictionRequest) (*couponServiceGrpc.CouponRestriction, error) {
	return nil, nil
}

func (h *Handler) GetCouponRestriction(ctx context.Context, in *couponServiceGrpc.GetCouponRestrictionRequest) (*couponServiceGrpc.CouponRestriction, error) {
	return nil, nil
}

func (h *Handler) UpdateCouponRestriction(ctx context.Context, in *couponServiceGrpc.UpdateCouponRestrictionRequest) (*couponServiceGrpc.CouponRestriction, error) {
	return nil, nil
}

func (h *Handler) DeleteCouponRestriction(ctx context.Context, in *couponServiceGrpc.DeleteCouponRestrictionRequest) (*couponServiceGrpc.DeleteCouponRestrictionResponse, error) {
	return nil, nil
}

func (h *Handler) ListCouponRestrictions(ctx context.Context, in *couponServiceGrpc.ListCouponRestrictionsRequest) (*couponServiceGrpc.ListCouponRestrictionsResponse, error) {
	return nil, nil
}

func (h *Handler) CreateUserCoupon(ctx context.Context, in *couponServiceGrpc.CreateUserCouponRequest) (*couponServiceGrpc.UserCoupon, error) {
	return nil, nil
}

func (h *Handler) GetUserCoupon(ctx context.Context, in *couponServiceGrpc.GetUserCouponRequest) (*couponServiceGrpc.UserCoupon, error) {
	return nil, nil
}

func (h *Handler) ListUserCoupons(ctx context.Context, in *couponServiceGrpc.ListUserCouponsRequest) (*couponServiceGrpc.ListUserCouponsResponse, error) {
	return nil, nil
}

func (h *Handler) UseUserCoupon(ctx context.Context, in *couponServiceGrpc.UseUserCouponRequest) (*couponServiceGrpc.UserCoupon, error) {
	return nil, nil
}

func (h *Handler) UnuseUserCoupon(ctx context.Context, in *couponServiceGrpc.UnuseUserCouponRequest) (*couponServiceGrpc.UserCoupon, error) {
	return nil, nil
}

func (h *Handler) ListCouponUsages(ctx context.Context, in *couponServiceGrpc.ListCouponUsagesRequest) (*couponServiceGrpc.ListCouponUsagesResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsages(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesRequest) (*couponServiceGrpc.ListUserCouponUsagesResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrder(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesByOrderRequest) (*couponServiceGrpc.ListUserCouponUsagesByOrderResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByCoupon(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesByCouponRequest) (*couponServiceGrpc.ListUserCouponUsagesByCouponResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrderAndCoupon(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesByOrderAndCouponRequest) (*couponServiceGrpc.ListUserCouponUsagesByOrderAndCouponResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrderAndCouponAndUser(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesByOrderAndCouponAndUserRequest) (*couponServiceGrpc.ListUserCouponUsagesByOrderAndCouponAndUserResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrderAndUser(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesByOrderAndUserRequest) (*couponServiceGrpc.ListUserCouponUsagesByOrderAndUserResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByCouponAndUser(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesByCouponAndUserRequest) (*couponServiceGrpc.ListUserCouponUsagesByCouponAndUserResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByUser(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesByUserRequest) (*couponServiceGrpc.ListUserCouponUsagesByUserResponse, error) {
	return nil, nil
}

func (h *Handler) ListUserCouponUsagesByOrderAndUserAndCoupon(ctx context.Context, in *couponServiceGrpc.ListUserCouponUsagesByOrderAndUserAndCouponRequest) (*couponServiceGrpc.ListUserCouponUsagesByOrderAndUserAndCouponResponse, error) {
	return nil, nil
}
