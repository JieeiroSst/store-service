package repository

import (
	"context"

	"github.com/JIeeiroSst/user-service/model"
	"gorm.io/gorm"
)

type RoleItemRepository struct {
	db *gorm.DB
}

type RoleItem interface {
	AddRoleItem(ctx context.Context, userRole model.RoleItem) error
	RemoveRoleItem(ctx context.Context, userId int) error
	UpdateRoleItem(ctx context.Context, userId int, roleId int) error
}

func NewRoleItemRepository(db *gorm.DB) *RoleItemRepository {
	return &RoleItemRepository{
		db: db,
	}
}

func (r *RoleItemRepository) AddRoleItem(ctx context.Context, userRole model.RoleItem) error {
	if err := r.db.Save(&userRole).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleItemRepository) RemoveRoleItem(ctx context.Context, userId int) error {
	if err := r.db.Delete(&model.RoleItem{}, "users_id = ?", userId).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleItemRepository) UpdateRoleItem(ctx context.Context, userId int, roleId int) error {
	if err := r.db.Model(&model.RoleItem{}).Where("users_id = ?", userId).Update("role_id", roleId).Error; err != nil {
		return err
	}
	return nil
}
