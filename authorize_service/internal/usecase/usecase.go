package usecase

import (
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/pkg/snowflake"
	"github.com/casbin/casbin/v2/persist"
)

type Usecase struct {
	Casbins Casbins
}

type Dependency struct {
	Repos     *repository.Repositories
	Snowflake snowflake.SnowflakeData
	Adapter   persist.Adapter
}

func NewUsecase(deps Dependency) *Usecase {
	casbinUsecase := NewCasbinUsecase(deps.Repos.Casbins, deps.Snowflake,deps.Adapter)

	return &Usecase{
		Casbins: casbinUsecase,
	}
}
