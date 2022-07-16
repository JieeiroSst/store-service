package repository

import "gorm.io/gorm"

type Repository struct {
	Users
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		Users: NewUserRepository(db),
	}
}
