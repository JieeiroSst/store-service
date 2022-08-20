package repository

import (
	"github.com/JIeeiroSst/user-service/model"
	"gorm.io/gorm"
)

type UserRoleRepository struct {
	db *gorm.DB
}

type UserRoles interface {
	AddRole(userRole model.UserRoles) error
	RemoveRole(userId int) error
	Update(userId int, roleId int) error
}

func NewUserRoleRepository(db *gorm.DB) *UserRoleRepository {
	return &UserRoleRepository{
		db: db,
	}
}

func (r *UserRoleRepository) AddRole(userRole model.UserRoles) error {
	if err := r.db.Save(&userRole).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserRoleRepository) RemoveRole(userId int) error {
	if err := r.db.Delete(&model.UserRoles{}, "users_id = ?", userId).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserRoleRepository) Update(userId int, roleId int) error {
	if err := r.db.Model(&model.UserRoles{}).Where("users_id = ?", userId).Update("role_id", roleId).Error; err != nil {
		return err
	}
	return nil
}
