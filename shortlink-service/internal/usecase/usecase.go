package usecase

import (
	"github.com/JIeeiroSst/shortlink-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Usecase struct {
	Links
}

type Dependency struct {
	Repos  *repository.Repositories
	Expire expire.CacheHelper
	Domain string
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Links: NewLinkUsecase(deps.Repos, deps.Expire, deps.Domain),
	}
}
