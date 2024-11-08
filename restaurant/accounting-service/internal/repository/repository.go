package repository

import "gorm.io/gorm"

type Repository struct {
	AuthCarts
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		AuthCarts: NewAuthCartRepository(db),
	}
}
