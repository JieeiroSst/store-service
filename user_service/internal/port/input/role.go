package input

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
)

type RoleService interface {
	GetRole(ctx context.Context, in dto.GetRoleRequest) (dto.GetRoleResponse, error)
	ListRoles(ctx context.Context, in dto.ListRolesRequest) (dto.ListRolesResponse, error)
	CreateRole(ctx context.Context, in dto.CreateRoleResquest) (dto.CreateRoleResponse, error)
	UpdateRole(ctx context.Context, in dto.UpdateRoleRequest) (dto.UpdateRoleResponse, error)
	DeleteRole(ctx context.Context, in dto.DeleteRoleRequest) (dto.DeleteRoleResponse, error)
}
