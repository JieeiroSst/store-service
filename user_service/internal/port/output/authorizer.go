package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type Authorizer interface {
	AssignRoles(ctx context.Context, userID int, roles ...string) (domain.UserRoles, error)
	SetUserRoles(ctx context.Context, userID int, roles ...string) (domain.UserRoles, error)
	RemoveUser(ctx context.Context, userID int) error
	GetUserRoles(ctx context.Context, userID int) (domain.UserRoles, error)
	RoleExists(ctx context.Context, role string) (bool, error)
}
