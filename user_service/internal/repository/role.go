package repository

import (
	"github.com/JIeeiroSst/user-service/common"
	"github.com/JIeeiroSst/user-service/model"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

type Roles interface {
	Create(role model.Role) error
	Update(id int, name string) error
	Delete(id int) error
	Role(id int) (*model.Role, error)
	Roles() ([]model.Role, error)
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

func (r *RoleRepository) Create(role model.Role) error {
	if err := r.db.Save(&role).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) Update(id int, name string) error {
	if err := r.db.Model(&model.Role{}).Where("id = ?", id).Update("name", name).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) Delete(id int) error {
	if err := r.db.Delete(&model.Role{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) Role(id int) (*model.Role, error) {
	var role model.Role
	query := r.db.Where("id =?", id).Preload("Users").Find(&role)
	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}
	if query.Error != nil {
		return nil, query.Error
	}
	return &role, nil
}

func (r *RoleRepository) Roles() ([]model.Role, error) {
	var roles []model.Role
	query := r.db.Preload("Users").Find(&roles)
	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}
	if query.Error != nil {
		return nil, query.Error
	}
	return roles, nil
}
