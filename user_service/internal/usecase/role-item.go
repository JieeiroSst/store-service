package usecase

import (
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/model"
)

type UserRoles interface {
	AddRole(userRole model.UserRoles) error
	RemoveRole(userId int) error
	Update(userId int, roleId int) error
}

type UserRoleUsecase struct {
	UserRoleRepo repository.UserRoles
}

func NewUserRoleUsecase(UserRoleRepo repository.UserRoles) *UserRoleUsecase {
	return &UserRoleUsecase{
		UserRoleRepo: UserRoleRepo,
	}
}

func (u *UserRoleUsecase) AddRole(userRole model.UserRoles) error {
	if err := u.UserRoleRepo.AddRole(userRole); err != nil {
		return err
	}
	return nil
}
func (u *UserRoleUsecase) RemoveRole(userId int) error {
	if err := u.UserRoleRepo.RemoveRole(userId); err != nil {
		return err
	}
	return nil
}

func (u *UserRoleUsecase) Update(userId int, roleId int) error {
	if err := u.UserRoleRepo.Update(userId, roleId); err != nil {
		return err
	}
	return nil
}
