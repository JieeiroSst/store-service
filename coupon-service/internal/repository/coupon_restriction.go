package repository

import "gorm.io/gorm"

type CouponRestrictionRepository interface {
}

type couponRestrictionRepository struct {
	db *gorm.DB
}

func NewCouponRestrictionRepository(db *gorm.DB) CouponRestrictionRepository {
	return &couponRestrictionRepository{db}
}
