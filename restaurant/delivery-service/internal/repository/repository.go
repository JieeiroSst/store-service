package repository

import "gorm.io/gorm"

type Repository struct {
	Deliveries
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Deliveries: NewDeliveryRepository(db),
	}
}
