package repository

import "gorm.io/gorm"

type UserAddress interface {
}

type UserAddressRepository struct {
	db *gorm.DB
}

func NewUserAddressRepository(db *gorm.DB) *UserAddressRepository {
	return &UserAddressRepository{
		db: db,
	}
}
