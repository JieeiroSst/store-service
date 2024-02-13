package usecase

import (
	"github.com/JIeeiroSst/oauth2-service/internal/repository"
	"github.com/JIeeiroSst/oauth2-service/pkg/cache"
)

type Usecase struct {
}

type Dependency struct {
	Repos       *repository.Repositories
	CacheHelper cache.CacheHelper
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{}
}
