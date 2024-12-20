package usecase

import (
	"github.com/JIeeiroSst/utils/cache/hash"
	"github.com/JIeerioSst/toggle-service/internal/repository"
)

type Countries interface {
}

type CountriesUsecase struct {
	repo    *repository.Repository
	configs *ConfigsUsecase
	cache   hash.CacheHelper
}

func NewCountriesUsecase(repo *repository.Repository,
	configs *ConfigsUsecase, cache hash.CacheHelper) *CountriesUsecase {
	return &CountriesUsecase{
		repo:    repo,
		configs: configs,
		cache:   cache,
	}
}
