package repository

import "gorm.io/gorm"

type Partneraddress interface {
}

type PartneraddressRepository struct {
	db *gorm.DB
}

func NewPartneraddressRepository(db *gorm.DB) *PartneraddressRepository {
	return &PartneraddressRepository{
		db: db,
	}
}
