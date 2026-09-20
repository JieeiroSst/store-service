package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
	"gorm.io/gorm"
)

type couponRestrictionRepository struct {
	db *gorm.DB
}

func NewCouponRestrictionRepository(db *gorm.DB) port.CouponRestrictionRepository {
	return &couponRestrictionRepository{db: db}
}

func (r *couponRestrictionRepository) Create(ctx context.Context, restriction *model.CouponRestriction) (*model.CouponRestriction, error) {
	if err := r.db.WithContext(ctx).Create(restriction).Error; err != nil {
		return nil, err
	}
	return restriction, nil
}

func (r *couponRestrictionRepository) GetByID(ctx context.Context, id int64) (*model.CouponRestriction, error) {
	var restriction model.CouponRestriction
	if err := r.db.WithContext(ctx).First(&restriction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &restriction, nil
}

func (r *couponRestrictionRepository) ListByCoupon(ctx context.Context, couponID int64) ([]model.CouponRestriction, error) {
	var restrictions []model.CouponRestriction
	err := r.db.WithContext(ctx).Where("coupon_id = ?", couponID).Find(&restrictions).Error
	return restrictions, err
}

func (r *couponRestrictionRepository) Update(ctx context.Context, restriction *model.CouponRestriction) (*model.CouponRestriction, error) {
	if err := r.db.WithContext(ctx).Save(restriction).Error; err != nil {
		return nil, err
	}
	return restriction, nil
}

func (r *couponRestrictionRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.CouponRestriction{}, id).Error
}
