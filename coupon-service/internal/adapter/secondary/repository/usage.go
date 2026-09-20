package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
	"gorm.io/gorm"
)

type couponUsageRepository struct {
	db *gorm.DB
}

func NewCouponUsageRepository(db *gorm.DB) port.CouponUsageRepository {
	return &couponUsageRepository{db: db}
}

func (r *couponUsageRepository) Create(ctx context.Context, usage *model.CouponUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *couponUsageRepository) GetByOrderAndCoupon(ctx context.Context, orderID, couponID int64) (*model.CouponUsage, error) {
	var usage model.CouponUsage
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND coupon_id = ?", orderID, couponID).
		First(&usage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &usage, nil
}

func (r *couponUsageRepository) ListByCoupon(ctx context.Context, couponID int64) ([]model.CouponUsage, error) {
	var usages []model.CouponUsage
	err := r.db.WithContext(ctx).Where("coupon_id = ?", couponID).Find(&usages).Error
	return usages, err
}

func (r *couponUsageRepository) ListByUser(ctx context.Context, userID int64) ([]model.CouponUsage, error) {
	var usages []model.CouponUsage
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&usages).Error
	return usages, err
}
