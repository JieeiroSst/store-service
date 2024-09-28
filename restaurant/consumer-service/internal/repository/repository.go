package repository

import "gorm.io/gorm"

type Repository struct {
	Consumers
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Consumers: NewConsumerRepository(db),
	}
}
