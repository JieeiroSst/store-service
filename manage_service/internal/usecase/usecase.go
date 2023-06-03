package usecase

import (
	"github.com/JIeeiroSst/manage-service/internal/repository"
	"github.com/JIeeiroSst/manage-service/pkg/cache"
	"github.com/JIeeiroSst/manage-service/pkg/snowflake"
)

type Usecase struct {
	UserKeyclock
}

type Dependency struct {
	CacheHelper cache.CacheHelper
	Repo        *repository.Repositories
	Snowflake   snowflake.SnowflakeData
}

func NewUsecase(deps Dependency) *Usecase {
	keyclockUsecase := NewUserKeycloakUsecase(deps.Repo.UserKeycloak)
	return &Usecase{
		UserKeyclock: keyclockUsecase,
	}
}
