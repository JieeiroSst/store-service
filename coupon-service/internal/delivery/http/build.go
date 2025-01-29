package http

import (
	pd "github.com/JIeeiroSst/coupon-service/gateway/proto"
	"github.com/JIeeiroSst/coupon-service/internal/dto"
)

func buildDtoCoupon(in *pd.CreateCouponRequest) dto.Coupon {
	return dto.Coupon{
		Code:              in.GetCode(),
		MinimumPurchase:   in.GetMinimumPurchase(),
		MaxDiscountAmount: in.GetMaxDiscountAmount(),
		Description:       in.GetDescription(),
		StartDate:         in.GetStartDate(),
		EndDate:           in.GetEndDate(),
		IsActive:          in.GetIsActive(),
		MaxUses:           in.GetMaxUses(),
		DiscountValue:     in.GetDiscountValue(),
		Type:              in.GetType(),
	}
}

func buildPbCoupon(coupon *dto.Coupon) pd.Coupon {
	return pd.Coupon{
		Code:              coupon.Code,
		MinimumPurchase:   coupon.MinimumPurchase,
		MaxDiscountAmount: coupon.MaxDiscountAmount,
		Description:       coupon.Description,
		StartDate:         coupon.StartDate,
		EndDate:           coupon.EndDate,
		IsActive:          coupon.IsActive,
		MaxUses:           coupon.MaxUses,
		DiscountValue:     coupon.DiscountValue,
		Type:              coupon.Type,
	}
}
