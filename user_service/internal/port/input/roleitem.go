package input

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
)

type RoleItemService interface {
	AddRoleItem(ctx context.Context, in dto.AddRoleItemRequest) (dto.AddRoleItemResponse, error)
	RemoveRoleItem(ctx context.Context, in dto.RemoveRoleItemRequest) (dto.RemoveRoleItemResponse, error)
	UpdateItemRole(ctx context.Context, in dto.UpdateRoleItemRequest) (dto.UpdateRoleItemResponse, error)
}
