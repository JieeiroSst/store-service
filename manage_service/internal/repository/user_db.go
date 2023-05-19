package repository

import "gorm.io/gorm"

type UserDB interface {
}

type UserDBRepo struct {
	db *gorm.DB
}

func NewUserDBRepo(db *gorm.DB) UserDB {
	return &UserDBRepo{
		db: db,
	}
}

