package repository

import "gorm.io/gorm"

type CouponUsageRepository interface {
}

type couponUsageRepository struct {
	db *gorm.DB
}

func NewCouponUsageRepository(db *gorm.DB) CouponUsageRepository {
	return &couponUsageRepository{db}
}
