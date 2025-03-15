package repository

import (
	"context"

	"github.com/JIeeiroSst/user-service/common"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/utils/pagination"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

type Roles interface {
	Create(ctx context.Context, role model.Role) error
	Update(ctx context.Context, id int, name string) error
	Delete(ctx context.Context, id int) error
	Role(ctx context.Context, id int) (*model.Role, error)
	Roles(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error)
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

func (r *RoleRepository) Create(ctx context.Context, role model.Role) error {
	if err := r.db.Save(&role).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) Update(ctx context.Context, id int, name string) error {
	if err := r.db.Model(&model.Role{}).Where("id = ?", id).Update("name", name).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.Delete(&model.Role{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) Role(ctx context.Context, id int) (*model.Role, error) {
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

func (r *RoleRepository) Roles(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error) {
	var roles []model.Role
	r.db.Scopes(pagination.Paginate(roles, &p, r.db)).Find(&roles)
	p.Rows = roles

	return p, nil
}
