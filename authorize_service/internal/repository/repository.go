package repository

import "gorm.io/gorm"

type Repositories struct {
	Casbins
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Casbins: NewCasbinRepo(db),
	}
}
