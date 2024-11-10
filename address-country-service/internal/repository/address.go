package repository

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/model"
	"gorm.io/gorm"
)

type AddressCountries interface {
	Save(ctx context.Context, address model.AddressCountry) error
}

type AddressCountryRepository struct {
	db *gorm.DB
}

func NewAddressCountryRepository(db *gorm.DB) *AddressCountryRepository {
	return &AddressCountryRepository{
		db: db,
	}
}

func (r *AddressCountryRepository) Save(ctx context.Context, address model.AddressCountry) error {
	if err := r.db.Create(&address).Error; err != nil {
		return err
	}
	return nil
}
