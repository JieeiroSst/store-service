package repository

import "gorm.io/gorm"

type Repository struct {
	Users
	Roles
	UserRoles
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		Users:     NewUserRepository(db),
		Roles:     NewRoleRepository(db),
		UserRoles: NewUserRoleRepository(db),
	}
}
