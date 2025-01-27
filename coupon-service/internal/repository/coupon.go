package repository

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/model"
	"gorm.io/gorm"
)

type CouponRepository interface {
	CreateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error)
}

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db}
}

func (r *couponRepository) CreateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error) {
	result := r.db.Create(&coupon)
	if result.Error != nil {
		return nil, result.Error
	}
	return coupon, nil
}
