package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
)

type userCouponService struct {
	userCoupons port.UserCouponRepository
	coupons     port.CouponRepository
}

func NewUserCouponService(userCoupons port.UserCouponRepository, coupons port.CouponRepository) port.UserCouponUsecase {
	return &userCouponService{userCoupons: userCoupons, coupons: coupons}
}

func (s *userCouponService) AssignCoupon(ctx context.Context, userID, couponID int64) (*model.UserCoupon, error) {
	if _, err := s.coupons.GetByID(ctx, couponID); err != nil {
		return nil, err
	}
	return s.userCoupons.Create(ctx, &model.UserCoupon{
		UserID:     userID,
		CouponID:   couponID,
		AssignedAt: time.Now(),
	})
}

func (s *userCouponService) ListByUser(ctx context.Context, userID int64) ([]model.UserCoupon, error) {
	return s.userCoupons.ListByUser(ctx, userID)
}

func (s *userCouponService) ListByCoupon(ctx context.Context, couponID int64) ([]model.UserCoupon, error) {
	return s.userCoupons.ListByCoupon(ctx, couponID)
}

func (s *userCouponService) UseCoupon(ctx context.Context, id int64) (*model.UserCoupon, error) {
	uc, err := s.userCoupons.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	uc.IsUsed = true
	uc.UsedAt = &now
	return s.userCoupons.Update(ctx, uc)
}

func (s *userCouponService) UnuseCoupon(ctx context.Context, id int64) (*model.UserCoupon, error) {
	uc, err := s.userCoupons.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	uc.IsUsed = false
	uc.UsedAt = nil
	return s.userCoupons.Update(ctx, uc)
}

func (s *userCouponService) DeleteUserCoupon(ctx context.Context, id int64) error {
	return s.userCoupons.Delete(ctx, id)
}
