package repository

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/model"
	"gorm.io/gorm"
)

type UserCouponRepository interface {
	GetByUserIDAndCouponID(ctx context.Context, userID, couponID uint) (*model.UserCoupon, error)
	Create(ctx context.Context, userCoupon *model.UserCoupon) (*model.UserCoupon, error)
	Update(ctx context.Context, userCoupon *model.UserCoupon) (*model.UserCoupon, error)
	Delete(ctx context.Context, userCoupon *model.UserCoupon) error
	GetByUserID(ctx context.Context, userID uint) ([]model.UserCoupon, error)
	GetByCouponID(ctx context.Context, couponID uint) ([]model.UserCoupon, error)
	GetUserCoupons(ctx context.Context) ([]model.UserCoupon, error)
}

type userCouponRepository struct {
	db *gorm.DB
}

func NewUserCouponRepository(db *gorm.DB) UserCouponRepository {
	return &userCouponRepository{db}
}

func (r *userCouponRepository) GetByUserIDAndCouponID(ctx context.Context, userID, couponID uint) (*model.UserCoupon, error) {
	var userCoupon model.UserCoupon
	if err := r.db.Where("user_id = ? AND coupon_id = ?", userID, couponID).First(&userCoupon).Error; err != nil {
		return nil, err
	}
	return &userCoupon, nil
}

func (r *userCouponRepository) Create(ctx context.Context, userCoupon *model.UserCoupon) (*model.UserCoupon, error) {
	if err := r.db.Create(userCoupon).Error; err != nil {
		return nil, err
	}
	return userCoupon, nil
}
func (r *userCouponRepository) Update(ctx context.Context, userCoupon *model.UserCoupon) (*model.UserCoupon, error) {
	if err := r.db.Save(userCoupon).Error; err != nil {
		return nil, err
	}
	return userCoupon, nil
}

func (r *userCouponRepository) Delete(ctx context.Context, userCoupon *model.UserCoupon) error {
	return r.db.Delete(userCoupon).Error
}

func (r *userCouponRepository) GetByUserID(ctx context.Context, userID uint) ([]model.UserCoupon, error) {
	var userCoupons []model.UserCoupon
	if err := r.db.Where("user_id = ?", userID).Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	return userCoupons, nil
}

func (r *userCouponRepository) GetByCouponID(ctx context.Context, couponID uint) ([]model.UserCoupon, error) {
	var userCoupons []model.UserCoupon
	if err := r.db.Where("coupon_id = ?", couponID).Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	return userCoupons, nil
}

func (r *userCouponRepository) GetUserCoupons(ctx context.Context) ([]model.UserCoupon, error) {
	var userCoupons []model.UserCoupon
	if err := r.db.Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	return userCoupons, nil
}
