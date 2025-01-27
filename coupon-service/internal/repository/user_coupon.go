package repository

import "gorm.io/gorm"

type UserCouponRepository interface{}

type userCouponRepository struct {
	db *gorm.DB
}

func NewUserCouponRepository(db *gorm.DB) UserCouponRepository {
	return &userCouponRepository{db}
}
