package repository

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/model"
	"gorm.io/gorm"
)

type CouponRepository interface {
	CreateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error)
	GetCoupon(ctx context.Context, id int) (*model.Coupon, error)
	GetCoupons(ctx context.Context) ([]*model.Coupon, error)
	UpdateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error)
	DeleteCoupon(ctx context.Context, id int) error
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

func (r *couponRepository) GetCoupon(ctx context.Context, id int) (*model.Coupon, error) {
	coupon := new(model.Coupon)
	result := r.db.First(coupon, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return coupon, nil
}

func (r *couponRepository) GetCoupons(ctx context.Context) ([]*model.Coupon, error) {
	var coupons []*model.Coupon
	result := r.db.Find(&coupons)
	if result.Error != nil {
		return nil, result.Error
	}
	return coupons, nil
}

func (r *couponRepository) UpdateCoupon(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error) {
	result := r.db.Save(coupon)
	if result.Error != nil {
		return nil, result.Error
	}
	return coupon, nil
}

func (r *couponRepository) DeleteCoupon(ctx context.Context, id int) error {
	result := r.db.Delete(&model.Coupon{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
