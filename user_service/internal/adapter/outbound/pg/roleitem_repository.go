package pg

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"gorm.io/gorm"
)

type RoleItemRepository struct {
	db *gorm.DB
}

func NewRoleItemRepository(db *gorm.DB) *RoleItemRepository {
	return &RoleItemRepository{db: db}
}

func (r *RoleItemRepository) AddRoleItem(ctx context.Context, roleItem domain.RoleItem) error {
	return r.db.Save(&roleItem).Error
}

func (r *RoleItemRepository) RemoveRoleItem(ctx context.Context, userId int) error {
	return r.db.Delete(&domain.RoleItem{}, "users_id = ?", userId).Error
}

func (r *RoleItemRepository) UpdateRoleItem(ctx context.Context, userId int, roleId int) error {
	return r.db.Model(&domain.RoleItem{}).Where("users_id = ?", userId).Update("role_id", roleId).Error
}
