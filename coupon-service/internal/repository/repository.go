package repository

import "gorm.io/gorm"

type Repositories struct {
	CouponRepository
	CouponRestrictionRepository
	CouponUsageRepository
	UserCouponRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		NewCouponRepository(db),
		NewCouponRestrictionRepository(db),
		NewCouponUsageRepository(db),
		NewUserCouponRepository(db),
	}
}
