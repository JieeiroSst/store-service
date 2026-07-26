package pg

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/utils/pagination"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(ctx context.Context, role domain.Role) error {
	return r.db.Save(&role).Error
}

func (r *RoleRepository) Update(ctx context.Context, id int, name string) error {
	return r.db.Model(&domain.Role{}).Where("id = ?", id).Update("name", name).Error
}

func (r *RoleRepository) Delete(ctx context.Context, id int) error {
	return r.db.Delete(&domain.Role{}, "id = ?", id).Error
}

func (r *RoleRepository) Role(ctx context.Context, id int) (*domain.Role, error) {
	var role domain.Role
	query := r.db.Where("id =?", id).Preload("Users").Find(&role)
	if query.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}
	if query.Error != nil {
		return nil, query.Error
	}
	return &role, nil
}

func (r *RoleRepository) Roles(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error) {
	var roles []domain.Role
	r.db.Scopes(pagination.Paginate(roles, &p, r.db)).Find(&roles)
	p.Rows = roles
	return p, nil
}
