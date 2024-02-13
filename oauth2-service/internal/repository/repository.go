package repository

import "gorm.io/gorm"

type Repositories struct {
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{}
}
