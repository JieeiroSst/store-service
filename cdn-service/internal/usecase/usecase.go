package usecase

import (
	"github.com/JIeeiroSst/cdn-service/config"
	"github.com/JIeeiroSst/cdn-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Usecase struct {
	CDNs
}

type Dependency struct {
	Repos    *repository.Repositories
	BaseHost config.BaseHostConfig
	Cache    expire.CacheHelper
}

func NewUsecase(deps Dependency) *Usecase {
	cdns := NewCDNUsecase(deps.Repos, deps.BaseHost, deps.Cache)
	return &Usecase{
		CDNs: cdns,
	}
}
