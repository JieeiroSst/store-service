package application

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
)

type restrictionService struct {
	restrictions port.CouponRestrictionRepository
}

func NewRestrictionService(restrictions port.CouponRestrictionRepository) port.CouponRestrictionUsecase {
	return &restrictionService{restrictions: restrictions}
}

func (s *restrictionService) CreateRestriction(ctx context.Context, restriction *model.CouponRestriction) (*model.CouponRestriction, error) {
	if !restriction.RestrictionType.Valid() {
		return nil, port.ErrInvalidRestrictionType
	}
	return s.restrictions.Create(ctx, restriction)
}

func (s *restrictionService) GetRestriction(ctx context.Context, id int64) (*model.CouponRestriction, error) {
	return s.restrictions.GetByID(ctx, id)
}

func (s *restrictionService) ListRestrictionsByCoupon(ctx context.Context, couponID int64) ([]model.CouponRestriction, error) {
	return s.restrictions.ListByCoupon(ctx, couponID)
}

func (s *restrictionService) UpdateRestriction(ctx context.Context, restriction *model.CouponRestriction) (*model.CouponRestriction, error) {
	if !restriction.RestrictionType.Valid() {
		return nil, port.ErrInvalidRestrictionType
	}
	return s.restrictions.Update(ctx, restriction)
}

func (s *restrictionService) DeleteRestriction(ctx context.Context, id int64) error {
	return s.restrictions.Delete(ctx, id)
}
