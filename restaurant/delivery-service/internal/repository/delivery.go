package repository

import "gorm.io/gorm"

type Deliveries interface {
	
}

type DeliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository {
	return &DeliveryRepository{
		db: db,
	}
}
