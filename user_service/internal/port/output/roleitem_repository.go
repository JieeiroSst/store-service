package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type RoleItemRepository interface {
	AddRoleItem(ctx context.Context, roleItem domain.RoleItem) error
	RemoveRoleItem(ctx context.Context, userID int) error
	UpdateRoleItem(ctx context.Context, userID int, roleID int) error
}
