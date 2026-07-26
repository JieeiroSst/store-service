package role

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/utils/copy"
	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/pagination"
)

type Service struct {
	roleRepo output.RoleRepository
}

func New(roleRepo output.RoleRepository) *Service {
	return &Service{roleRepo: roleRepo}
}

func (s *Service) CreateRole(ctx context.Context, in dto.CreateRoleResquest) (dto.CreateRoleResponse, error) {
	role := domain.Role{
		Id:   geared_id.GearedIntID(),
		Name: in.Name,
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return dto.CreateRoleResponse{}, err
	}
	return dto.CreateRoleResponse{}, nil
}

func (s *Service) UpdateRole(ctx context.Context, in dto.UpdateRoleRequest) (dto.UpdateRoleResponse, error) {
	if err := s.roleRepo.Update(ctx, int(in.Id), in.Name); err != nil {
		return dto.UpdateRoleResponse{}, err
	}
	return dto.UpdateRoleResponse{}, nil
}

func (s *Service) DeleteRole(ctx context.Context, in dto.DeleteRoleRequest) (dto.DeleteRoleResponse, error) {
	if err := s.roleRepo.Delete(ctx, int(in.Id)); err != nil {
		return dto.DeleteRoleResponse{}, err
	}
	return dto.DeleteRoleResponse{}, nil
}

func (s *Service) GetRole(ctx context.Context, in dto.GetRoleRequest) (dto.GetRoleResponse, error) {
	role, err := s.roleRepo.Role(ctx, int(in.Id))
	if err != nil {
		return dto.GetRoleResponse{}, err
	}

	var res dto.GetRoleResponse
	if err := copy.CopyObject(&role, &res.Role); err != nil {
		return dto.GetRoleResponse{}, err
	}
	return res, nil
}

func (s *Service) ListRoles(ctx context.Context, in dto.ListRolesRequest) (dto.ListRolesResponse, error) {
	var p pagination.Pagination
	if err := copy.CopyObject(&in, &p); err != nil {
		return dto.ListRolesResponse{}, err
	}
	roles, err := s.roleRepo.Roles(ctx, p)
	if err != nil {
		return dto.ListRolesResponse{}, err
	}

	var res dto.ListRolesResponse
	if err := copy.CopyObject(&roles, &res.Roles); err != nil {
		return dto.ListRolesResponse{}, err
	}
	return res, nil
}
