package repository

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/model"
	"gorm.io/gorm"
)

type CouponUsageRepository interface {
	CreateCouponUsage(ctx context.Context, couponUsage model.CouponUsage) error
	GetCouponUsage(ctx context.Context, couponID int) (*model.CouponUsage, error)
	GetCouponUsages(ctx context.Context) ([]*model.CouponUsage, error)
	UpdateCouponUsage(ctx context.Context, couponUsage model.CouponUsage) error
	DeleteCouponUsage(ctx context.Context, couponID int) error
}

type couponUsageRepository struct {
	db *gorm.DB
}

func NewCouponUsageRepository(db *gorm.DB) CouponUsageRepository {
	return &couponUsageRepository{db}
}

func (r *couponUsageRepository) CreateCouponUsage(ctx context.Context, couponUsage model.CouponUsage) error {
	result := r.db.Create(&couponUsage)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *couponUsageRepository) GetCouponUsage(ctx context.Context, couponID int) (*model.CouponUsage, error) {
	couponUsage := new(model.CouponUsage)
	result := r.db.First(couponUsage, couponID)
	if result.Error != nil {
		return nil, result.Error
	}
	return couponUsage, nil
}

func (r *couponUsageRepository) GetCouponUsages(ctx context.Context) ([]*model.CouponUsage, error) {
	var couponUsages []*model.CouponUsage
	result := r.db.Find(&couponUsages)
	if result.Error != nil {
		return nil, result.Error
	}
	return couponUsages, nil
}

func (r *couponUsageRepository) UpdateCouponUsage(ctx context.Context, couponUsage model.CouponUsage) error {
	result := r.db.Save(couponUsage)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *couponUsageRepository) DeleteCouponUsage(ctx context.Context, couponID int) error {
	result := r.db.Delete(&model.CouponUsage{}, couponID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}