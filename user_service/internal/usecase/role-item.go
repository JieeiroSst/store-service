package usecase

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/utils/copy"
)

type RoleItem interface {
	AddRoleItem(ctx context.Context, in dto.AddRoleItemRequest) (dto.AddRoleItemResponse, error)
	RemoveRoleItem(ctx context.Context, in dto.RemoveRoleItemRequest) (dto.RemoveRoleItemResponse, error)
	UpdateItemRole(ctx context.Context, in dto.UpdateRoleItemRequest) (dto.UpdateRoleItemResponse, error)
}

type UserRoleUsecase struct {
	UserRoleRepo repository.RoleItem
}

func NewUserRoleUsecase(UserRoleRepo repository.RoleItem) *UserRoleUsecase {
	return &UserRoleUsecase{
		UserRoleRepo: UserRoleRepo,
	}
}

func (u *UserRoleUsecase) AddRoleItem(ctx context.Context, in dto.AddRoleItemRequest) (dto.AddRoleItemResponse, error) {
	var userRole model.RoleItem
	if err := copy.CopyObject(&in, &userRole); err != nil {
		return dto.AddRoleItemResponse{}, err
	}
	if err := u.UserRoleRepo.AddRoleItem(ctx, userRole); err != nil {
		return dto.AddRoleItemResponse{}, err
	}
	var res dto.AddRoleItemResponse
	res.Message = "success"
	return res, nil
}
func (u *UserRoleUsecase) RemoveRoleItem(ctx context.Context, in dto.RemoveRoleItemRequest) (dto.RemoveRoleItemResponse, error) {
	if err := u.UserRoleRepo.RemoveRoleItem(ctx, int(in.UserId)); err != nil {
		return dto.RemoveRoleItemResponse{Message: "failed"}, err
	}
	var res dto.RemoveRoleItemResponse
	res.Message = "success"
	return res, nil
}

func (u *UserRoleUsecase) UpdateItemRole(ctx context.Context, in dto.UpdateRoleItemRequest) (dto.UpdateRoleItemResponse, error) {
	if err := u.UserRoleRepo.UpdateRoleItem(ctx, int(in.UserId), int(in.RoleId)); err != nil {
		return dto.UpdateRoleItemResponse{Message: "failed"}, err
	}
	return dto.UpdateRoleItemResponse{Message: "success"}, nil
}
