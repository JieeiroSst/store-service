package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
	"gorm.io/gorm"
)

type userCouponRepository struct {
	db *gorm.DB
}

func NewUserCouponRepository(db *gorm.DB) port.UserCouponRepository {
	return &userCouponRepository{db: db}
}

func (r *userCouponRepository) Create(ctx context.Context, userCoupon *model.UserCoupon) (*model.UserCoupon, error) {
	if err := r.db.WithContext(ctx).Create(userCoupon).Error; err != nil {
		return nil, err
	}
	return userCoupon, nil
}

func (r *userCouponRepository) GetByID(ctx context.Context, id int64) (*model.UserCoupon, error) {
	var userCoupon model.UserCoupon
	if err := r.db.WithContext(ctx).First(&userCoupon, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &userCoupon, nil
}

func (r *userCouponRepository) GetByUserAndCoupon(ctx context.Context, userID, couponID int64) (*model.UserCoupon, error) {
	var userCoupon model.UserCoupon
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		First(&userCoupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &userCoupon, nil
}

func (r *userCouponRepository) Update(ctx context.Context, userCoupon *model.UserCoupon) (*model.UserCoupon, error) {
	if err := r.db.WithContext(ctx).Save(userCoupon).Error; err != nil {
		return nil, err
	}
	return userCoupon, nil
}

func (r *userCouponRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.UserCoupon{}, id).Error
}

func (r *userCouponRepository) ListByUser(ctx context.Context, userID int64) ([]model.UserCoupon, error) {
	var userCoupons []model.UserCoupon
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userCoupons).Error
	return userCoupons, err
}

func (r *userCouponRepository) ListByCoupon(ctx context.Context, couponID int64) ([]model.UserCoupon, error) {
	var userCoupons []model.UserCoupon
	err := r.db.WithContext(ctx).Where("coupon_id = ?", couponID).Find(&userCoupons).Error
	return userCoupons, err
}
