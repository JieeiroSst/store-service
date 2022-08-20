package usecase

import (
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/user-service/pkg/snowflake"
)

type Roles interface {
	Create(role model.Role) error
	Update(id int, name string) error
	Delete(id int) error
	Role(id int) (*model.Role, error)
	Roles() ([]model.Role, error)
}

type RoleUsecase struct {
	RoleRepo  repository.Roles
	Snowflake snowflake.SnowflakeData
}

func NewRoleUsecase(RoleRepo repository.Roles,
	Snowflake snowflake.SnowflakeData) *RoleUsecase {
	return &RoleUsecase{
		RoleRepo:  RoleRepo,
		Snowflake: Snowflake,
	}
}

func (u *RoleUsecase) Create(req model.Role) error {
	role := model.Role{
		Id:   u.Snowflake.GearedID(),
		Name: req.Name,
	}
	if err := u.RoleRepo.Create(role); err != nil {
		return err
	}
	return nil
}

func (u *RoleUsecase) Update(id int, name string) error {
	if err := u.RoleRepo.Update(id, name); err != nil {
		return err
	}
	return nil
}

func (u *RoleUsecase) Delete(id int) error {
	if err := u.RoleRepo.Delete(id); err != nil {
		return err
	}
	return nil
}

func (u *RoleUsecase) Role(id int) (*model.Role, error) {
	role, err := u.RoleRepo.Role(id)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (u *RoleUsecase) Roles() ([]model.Role, error) {
	roles, err := u.RoleRepo.Roles()
	if err != nil {
		return nil, err
	}
	return roles, nil
}
