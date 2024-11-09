package repository

import "gorm.io/gorm"

type OrderShippingaddress interface {
}

type OrderShippingaddressRepository struct {
	db *gorm.DB
}

func NewOrderShippingaddressRepository(db *gorm.DB) *OrderShippingaddressRepository {
	return &OrderShippingaddressRepository{
		db: db,
	}
}
