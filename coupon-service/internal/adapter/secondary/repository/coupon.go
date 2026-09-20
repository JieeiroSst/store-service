package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
	"gorm.io/gorm"
)

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) port.CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) Create(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error) {
	if err := r.db.WithContext(ctx).Create(coupon).Error; err != nil {
		return nil, err
	}
	return coupon, nil
}

func (r *couponRepository) GetByID(ctx context.Context, id int64) (*model.Coupon, error) {
	var coupon model.Coupon
	if err := r.db.WithContext(ctx).First(&coupon, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) GetByCode(ctx context.Context, code string) (*model.Coupon, error) {
	var coupon model.Coupon
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) List(ctx context.Context) ([]model.Coupon, error) {
	var coupons []model.Coupon
	err := r.db.WithContext(ctx).Order("id").Find(&coupons).Error
	return coupons, err
}

func (r *couponRepository) Update(ctx context.Context, coupon *model.Coupon) (*model.Coupon, error) {
	if err := r.db.WithContext(ctx).Save(coupon).Error; err != nil {
		return nil, err
	}
	return coupon, nil
}

func (r *couponRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Coupon{}, id).Error
}

func (r *couponRepository) IncrementUsage(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("id = ? AND (max_uses IS NULL OR current_uses < max_uses)", id).
		UpdateColumn("current_uses", gorm.Expr("current_uses + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrUsageLimitReached
	}
	return nil
}

func (r *couponRepository) DeactivateStale(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("is_active = true AND (end_date < ? OR (max_uses IS NOT NULL AND current_uses >= max_uses))", now).
		Update("is_active", false)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
