package repository

import (
	"context"

	"github.com/JIeeiroSst/coupon-service/internal/model"
	"gorm.io/gorm"
)

type CouponRestrictionRepository interface {
	CreateCouponRestriction(ctx context.Context, couponRestriction *model.CouponRestriction) (*model.CouponRestriction, error)
	GetCouponRestriction(ctx context.Context, id int) (*model.CouponRestriction, error)
	GetCouponRestrictions(ctx context.Context) ([]*model.CouponRestriction, error)
	UpdateCouponRestriction(ctx context.Context, couponRestriction *model.CouponRestriction) (*model.CouponRestriction, error)
	DeleteCouponRestriction(ctx context.Context, id int) error
}

type couponRestrictionRepository struct {
	db *gorm.DB
}

func NewCouponRestrictionRepository(db *gorm.DB) CouponRestrictionRepository {
	return &couponRestrictionRepository{db}
}

func (r *couponRestrictionRepository) CreateCouponRestriction(ctx context.Context, couponRestriction *model.CouponRestriction) (*model.CouponRestriction, error) {
	result := r.db.Create(&couponRestriction)
	if result.Error != nil {
		return nil, result.Error
	}
	return couponRestriction, nil
}

func (r *couponRestrictionRepository) GetCouponRestriction(ctx context.Context, id int) (*model.CouponRestriction, error) {
	couponRestriction := new(model.CouponRestriction)
	result := r.db.First(couponRestriction, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return couponRestriction, nil
}

func (r *couponRestrictionRepository) GetCouponRestrictions(ctx context.Context) ([]*model.CouponRestriction, error) {
	var couponRestrictions []*model.CouponRestriction
	result := r.db.Find(&couponRestrictions)
	if result.Error != nil {
		return nil, result.Error
	}
	return couponRestrictions, nil
}

func (r *couponRestrictionRepository) UpdateCouponRestriction(ctx context.Context, couponRestriction *model.CouponRestriction) (*model.CouponRestriction, error) {
	result := r.db.Save(couponRestriction)
	if result.Error != nil {
		return nil, result.Error
	}
	return couponRestriction, nil
}

func (r *couponRestrictionRepository) DeleteCouponRestriction(ctx context.Context, id int) error {
	result := r.db.Delete(&model.CouponRestriction{}, id)
	return result.Error
}
