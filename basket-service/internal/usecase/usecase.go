package usecase

import (
	"github.com/JIeeiroSst/basket-service/internal/repository"
	"github.com/JIeeiroSst/basket-service/pkg/cache"
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
