package usecase

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/utils/copy"
	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/pagination"
)

type Roles interface {
	GetRole(ctx context.Context, in dto.GetRoleRequest) (dto.GetRoleResponse, error)
	ListRoles(ctx context.Context, in dto.ListRolesRequest) (dto.ListRolesResponse, error)
	CreateRole(ctx context.Context, in dto.CreateRoleResquest) (dto.CreateRoleResponse, error)
	UpdateRole(ctx context.Context, in dto.UpdateRoleRequest) (dto.UpdateRoleResponse, error)
	DeleteRole(ctx context.Context, in dto.DeleteRoleRequest) (dto.DeleteRoleResponse, error)
}

type RoleUsecase struct {
	RoleRepo repository.Roles
}

func NewRoleUsecase(RoleRepo repository.Roles) *RoleUsecase {
	return &RoleUsecase{
		RoleRepo: RoleRepo,
	}
}

func (u *RoleUsecase) CreateRole(ctx context.Context, in dto.CreateRoleResquest) (dto.CreateRoleResponse, error) {
	role := model.Role{
		Id:   geared_id.GearedIntID(),
		Name: in.Name,
	}
	if err := u.RoleRepo.Create(ctx, role); err != nil {
		return dto.CreateRoleResponse{}, err
	}
	return dto.CreateRoleResponse{}, nil
}

func (u *RoleUsecase) UpdateRole(ctx context.Context, in dto.UpdateRoleRequest) (dto.UpdateRoleResponse, error) {
	if err := u.RoleRepo.Update(ctx, int(in.Id), in.Name); err != nil {
		return dto.UpdateRoleResponse{}, err
	}
	return dto.UpdateRoleResponse{}, nil
}

func (u *RoleUsecase) DeleteRole(ctx context.Context, in dto.DeleteRoleRequest) (dto.DeleteRoleResponse, error) {
	if err := u.RoleRepo.Delete(ctx, int(in.Id)); err != nil {
		return dto.DeleteRoleResponse{}, err
	}
	return dto.DeleteRoleResponse{}, nil
}

func (u *RoleUsecase) GetRole(ctx context.Context, in dto.GetRoleRequest) (dto.GetRoleResponse, error) {
	role, err := u.RoleRepo.Role(ctx, int(in.Id))
	if err != nil {
		return dto.GetRoleResponse{}, err
	}

	var res dto.GetRoleResponse
	if err := copy.CopyObject(&role, &res.Role); err != nil {
		return dto.GetRoleResponse{}, err
	}
	return res, nil
}

func (u *RoleUsecase) ListRoles(ctx context.Context, in dto.ListRolesRequest) (dto.ListRolesResponse, error) {
	var p pagination.Pagination
	if err := copy.CopyObject(&in, &p); err != nil {
		return dto.ListRolesResponse{}, err
	}
	roles, err := u.RoleRepo.Roles(ctx, p)
	if err != nil {
		return dto.ListRolesResponse{}, err
	}

	var res dto.ListRolesResponse
	if err := copy.CopyObject(&roles, &res.Roles); err != nil {
		return dto.ListRolesResponse{}, err
	}
	return res, nil
}
