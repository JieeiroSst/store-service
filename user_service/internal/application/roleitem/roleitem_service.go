package roleitem

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/utils/copy"
)

type Service struct {
	roleItemRepo output.RoleItemRepository
}

func New(roleItemRepo output.RoleItemRepository) *Service {
	return &Service{roleItemRepo: roleItemRepo}
}

func (s *Service) AddRoleItem(ctx context.Context, in dto.AddRoleItemRequest) (dto.AddRoleItemResponse, error) {
	var roleItem domain.RoleItem
	if err := copy.CopyObject(&in, &roleItem); err != nil {
		return dto.AddRoleItemResponse{}, err
	}
	if err := s.roleItemRepo.AddRoleItem(ctx, roleItem); err != nil {
		return dto.AddRoleItemResponse{}, err
	}
	return dto.AddRoleItemResponse{Message: "success"}, nil
}

func (s *Service) RemoveRoleItem(ctx context.Context, in dto.RemoveRoleItemRequest) (dto.RemoveRoleItemResponse, error) {
	if err := s.roleItemRepo.RemoveRoleItem(ctx, int(in.UserId)); err != nil {
		return dto.RemoveRoleItemResponse{Message: "failed"}, err
	}
	return dto.RemoveRoleItemResponse{Message: "success"}, nil
}

func (s *Service) UpdateItemRole(ctx context.Context, in dto.UpdateRoleItemRequest) (dto.UpdateRoleItemResponse, error) {
	if err := s.roleItemRepo.UpdateRoleItem(ctx, int(in.UserId), int(in.RoleId)); err != nil {
		return dto.UpdateRoleItemResponse{Message: "failed"}, err
	}
	return dto.UpdateRoleItemResponse{Message: "success"}, nil
}
