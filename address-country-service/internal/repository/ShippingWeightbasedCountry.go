package repository

import "gorm.io/gorm"

type ShippingWeightbasedCountries interface {
}

type ShippingWeightbasedCountryRepository struct {
	db *gorm.DB
}

func NewShippingWeightbasedCountryRepository(db *gorm.DB) *ShippingWeightbasedCountryRepository {
	return &ShippingWeightbasedCountryRepository{
		db: db,
	}
}
