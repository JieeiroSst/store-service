package repository

import "gorm.io/gorm"

type Repository struct {
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{}
}
