package application

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
)

type usageService struct {
	usages port.CouponUsageRepository
}

func NewUsageService(usages port.CouponUsageRepository) port.CouponUsageUsecase {
	return &usageService{usages: usages}
}

func (s *usageService) ListUsagesByCoupon(ctx context.Context, couponID int64) ([]model.CouponUsage, error) {
	return s.usages.ListByCoupon(ctx, couponID)
}

func (s *usageService) ListUsagesByUser(ctx context.Context, userID int64) ([]model.CouponUsage, error) {
	return s.usages.ListByUser(ctx, userID)
}
