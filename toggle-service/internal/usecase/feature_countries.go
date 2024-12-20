package usecase

import (
	"github.com/JIeeiroSst/utils/cache/hash"
	"github.com/JIeerioSst/toggle-service/internal/repository"
)

type FeatureCountries interface {
}

type FeatureCountriesUsecase struct {
	repo  *repository.Repository
	cache hash.CacheHelper
}

func NewFeatureCountriesUsecase(repo *repository.Repository,
	cache hash.CacheHelper) *FeatureCountriesUsecase {
	return &FeatureCountriesUsecase{
		repo:  repo,
		cache: cache,
	}
}
