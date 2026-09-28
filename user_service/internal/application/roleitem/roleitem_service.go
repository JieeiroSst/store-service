package roleitem

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/port/output"
)

type Service struct {
	roleRepo   output.RoleRepository
	authorizer output.Authorizer
}

func New(roleRepo output.RoleRepository, authorizer output.Authorizer) *Service {
	return &Service{roleRepo: roleRepo, authorizer: authorizer}
}

func (s *Service) AddRoleItem(ctx context.Context, in dto.AddRoleItemRequest) (dto.AddRoleItemResponse, error) {
	role, err := s.roleRepo.Role(ctx, int(in.RoleId))
	if err != nil {
		return dto.AddRoleItemResponse{Message: "failed"}, err
	}
	if _, err := s.authorizer.AssignRoles(ctx, int(in.UserId), role.Name); err != nil {
		return dto.AddRoleItemResponse{Message: "failed"}, err
	}
	return dto.AddRoleItemResponse{
		Role:    &dto.Role{Id: int32(role.Id), Name: role.Name},
		Message: "success",
	}, nil
}

func (s *Service) RemoveRoleItem(ctx context.Context, in dto.RemoveRoleItemRequest) (dto.RemoveRoleItemResponse, error) {
	if err := s.authorizer.RemoveUser(ctx, int(in.UserId)); err != nil {
		return dto.RemoveRoleItemResponse{Message: "failed"}, err
	}
	return dto.RemoveRoleItemResponse{Message: "success"}, nil
}

func (s *Service) UpdateItemRole(ctx context.Context, in dto.UpdateRoleItemRequest) (dto.UpdateRoleItemResponse, error) {
	role, err := s.roleRepo.Role(ctx, int(in.RoleId))
	if err != nil {
		return dto.UpdateRoleItemResponse{Message: "failed"}, err
	}
	if _, err := s.authorizer.SetUserRoles(ctx, int(in.UserId), role.Name); err != nil {
		return dto.UpdateRoleItemResponse{Message: "failed"}, err
	}
	return dto.UpdateRoleItemResponse{
		Role:    &dto.Role{Id: int32(role.Id), Name: role.Name},
		Message: "success",
	}, nil
}
