package repository

import "gorm.io/gorm"

type Repository struct {
	Orders
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		Orders: NewOrderRepository(db),
	}
}
