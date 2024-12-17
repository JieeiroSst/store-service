package usecase

import (
	"github.com/JIeeiroSst/media-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Usecase struct {
}

type Dependency struct {
	Repos       *repository.Repository
	CacheHelper expire.CacheHelper
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{}
}
