package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/utils/pagination"
)

type RoleRepository interface {
	Create(ctx context.Context, role domain.Role) error
	Update(ctx context.Context, id int, name string) error
	Delete(ctx context.Context, id int) error
	Role(ctx context.Context, id int) (*domain.Role, error)
	Roles(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error)
}
