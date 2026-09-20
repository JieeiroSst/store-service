package application

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
)

type couponService struct {
	coupons      port.CouponRepository
	restrictions port.CouponRestrictionRepository
	usages       port.CouponUsageRepository
	userCoupons  port.UserCouponRepository
}

func NewCouponService(
	coupons port.CouponRepository,
	restrictions port.CouponRestrictionRepository,
	usages port.CouponUsageRepository,
	userCoupons port.UserCouponRepository,
) port.CouponUsecase {
	return &couponService{coupons: coupons, restrictions: restrictions, usages: usages, userCoupons: userCoupons}
}

func validateCoupon(coupon *model.Coupon) error {
	if !coupon.Type.Valid() {
		return port.ErrInvalidCouponType
	}
	if !coupon.EndDate.After(coupon.StartDate) {
		return port.ErrInvalidDateRange
	}
	return nil
}

func (s *couponService) CreateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error) {
	if err := validateCoupon(coupon); err != nil {
		return nil, err
	}
	return s.coupons.Create(ctx, coupon)
}

func (s *couponService) GetCoupon(ctx context.Context, id int64) (*model.Coupon, error) {
	return s.coupons.GetByID(ctx, id)
}

func (s *couponService) GetCouponByCode(ctx context.Context, code string) (*model.Coupon, error) {
	return s.coupons.GetByCode(ctx, code)
}

func (s *couponService) ListCoupons(ctx context.Context) ([]model.Coupon, error) {
	return s.coupons.List(ctx)
}

func (s *couponService) UpdateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error) {
	if err := validateCoupon(coupon); err != nil {
		return nil, err
	}
	return s.coupons.Update(ctx, coupon)
}

func (s *couponService) DeleteCoupon(ctx context.Context, id int64) error {
	return s.coupons.Delete(ctx, id)
}

func (s *couponService) ValidateCoupon(ctx context.Context, in port.ValidateCouponInput) (*model.Coupon, float64, error) {
	coupon, err := s.coupons.GetByCode(ctx, in.Code)
	if err != nil {
		return nil, 0, err
	}

	if !coupon.IsActive {
		return nil, 0, port.ErrCouponInactive
	}
	if !coupon.IsWithinWindow(time.Now()) {
		return nil, 0, port.ErrCouponExpired
	}
	if in.PurchaseAmount < coupon.MinimumPurchase {
		return nil, 0, port.ErrMinimumPurchaseNotMet
	}
	if !coupon.HasUsesLeft() {
		return nil, 0, port.ErrUsageLimitReached
	}

	restrictions, err := s.restrictions.ListByCoupon(ctx, coupon.ID)
	if err != nil {
		return nil, 0, err
	}
	if err := checkRestrictions(restrictions, in.CategoryIDs, in.ProductIDs); err != nil {
		return nil, 0, err
	}

	if err := s.checkUserEligibility(ctx, coupon.ID, in.UserID); err != nil {
		return nil, 0, err
	}

	return coupon, coupon.CalculateDiscount(in.PurchaseAmount), nil
}

func checkRestrictions(restrictions []model.CouponRestriction, categoryIDs, productIDs []int64) error {
	if err := checkRestrictionsByType(restrictions, model.RestrictionCategory, categoryIDs); err != nil {
		return err
	}
	return checkRestrictionsByType(restrictions, model.RestrictionProduct, productIDs)
}

func checkRestrictionsByType(restrictions []model.CouponRestriction, restrictionType model.RestrictionType, entityIDs []int64) error {
	var includes, excludes []int64
	for _, r := range restrictions {
		if r.RestrictionType != restrictionType {
			continue
		}
		if r.IsExclude {
			excludes = append(excludes, r.RestrictedEntityID)
		} else {
			includes = append(includes, r.RestrictedEntityID)
		}
	}

	for _, id := range entityIDs {
		if contains(excludes, id) {
			return port.ErrCouponRestricted
		}
	}

	if len(includes) > 0 && !anyContains(includes, entityIDs) {
		return port.ErrCouponRestricted
	}

	return nil
}

func contains(ids []int64, id int64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func anyContains(haystack, needles []int64) bool {
	for _, id := range needles {
		if contains(haystack, id) {
			return true
		}
	}
	return false
}

func (s *couponService) checkUserEligibility(ctx context.Context, couponID, userID int64) error {
	targeted, err := s.userCoupons.ListByCoupon(ctx, couponID)
	if err != nil {
		return err
	}
	if len(targeted) == 0 {
		return nil
	}

	uc, err := s.userCoupons.GetByUserAndCoupon(ctx, userID, couponID)
	if err != nil {
		if errors.Is(err, port.ErrNotFound) {
			return port.ErrCouponNotAssignedToUser
		}
		return err
	}
	if uc.IsUsed {
		return port.ErrCouponAlreadyUsedByUser
	}
	return nil
}

func (s *couponService) ApplyCoupon(ctx context.Context, in port.ApplyCouponInput) (*model.CouponUsage, error) {
	coupon, discount, err := s.ValidateCoupon(ctx, in.ValidateCouponInput)
	if err != nil {
		return nil, err
	}

	if _, err := s.usages.GetByOrderAndCoupon(ctx, in.OrderID, coupon.ID); err == nil {
		return nil, port.ErrOrderAlreadyUsedCoupon
	} else if !errors.Is(err, port.ErrNotFound) {
		return nil, err
	}

	if err := s.coupons.IncrementUsage(ctx, coupon.ID); err != nil {
		return nil, err
	}

	usage := &model.CouponUsage{
		CouponID:       coupon.ID,
		UserID:         in.UserID,
		OrderID:        in.OrderID,
		DiscountAmount: discount,
		UsedAt:         time.Now(),
	}
	if err := s.usages.Create(ctx, usage); err != nil {
		return nil, err
	}

	if uc, err := s.userCoupons.GetByUserAndCoupon(ctx, in.UserID, coupon.ID); err == nil {
		now := time.Now()
		uc.IsUsed = true
		uc.UsedAt = &now
		_, _ = s.userCoupons.Update(ctx, uc)
	}

	return usage, nil
}

func (s *couponService) DeactivateStaleCoupons(ctx context.Context) (int64, error) {
	return s.coupons.DeactivateStale(ctx, time.Now())
}
