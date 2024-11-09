package repository

import "gorm.io/gorm"

type ShippingOrderanditemchangesCountry interface {
}

type ShippingOrderanditemchangesCountryRepository struct {
	db *gorm.DB
}

func NewShippingOrderanditemchangesCountryRepository(db *gorm.DB) *ShippingOrderanditemchangesCountryRepository {
	return &ShippingOrderanditemchangesCountryRepository{
		db: db,
	}
}

