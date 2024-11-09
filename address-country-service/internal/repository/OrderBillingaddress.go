package repository

import "gorm.io/gorm"

type OrderBillingaddress interface {
}

type OrderBillingaddressRepositorry struct {
	db *gorm.DB
}

func NewOrderBillingaddressRepositorry(db *gorm.DB) *OrderBillingaddressRepositorry {
	return &OrderBillingaddressRepositorry{
		db: db,
	}
}
