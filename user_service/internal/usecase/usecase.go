package usecase

import (
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/token"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Usecase struct {
	Users
	Roles
	RoleItem
}

type Dependency struct {
	Repos *repository.Repository
	Hash  hash.Hash
	Token token.Tokens
	Cache expire.CacheHelper
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Users:    NewUsercase(deps.Repos.Users, deps.Hash, deps.Token),
		Roles:    NewRoleUsecase(deps.Repos.Roles),
		RoleItem: NewUserRoleUsecase(deps.Repos.RoleItem),
	}
}
