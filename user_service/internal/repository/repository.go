package repository

import "gorm.io/gorm"

type Repository struct {
	Users
	Roles
	RoleItem
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		Users:    NewUserRepository(db),
		Roles:    NewRoleRepository(db),
		RoleItem: NewRoleItemRepository(db),
	}
}
