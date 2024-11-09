package repository

import "gorm.io/gorm"

type AddressCountries interface {
}

type AddressCountryRepository struct {
	db *gorm.DB
}

func NewAddressCountryRepository(db *gorm.DB) *AddressCountryRepository {
	return &AddressCountryRepository{
		db: db,
	}
}
