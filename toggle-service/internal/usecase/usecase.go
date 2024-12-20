package usecase

import (
	"github.com/JIeeiroSst/utils/cache/hash"
	"github.com/JIeerioSst/toggle-service/internal/repository"
)

type Usecase struct {
	Configs
	Countries
	FeatureCountries
}

type Dependency struct {
	Repos           *repository.Repository
	CacheHelper     hash.CacheHelper
	UnidocSerectKey string
}

func NewUsecase(deps Dependency) *Usecase {
	config := NewConfigsUsecase()
	countries := NewCountriesUsecase(deps.Repos, config, deps.CacheHelper)
	featureCountries := NewFeatureCountriesUsecase(deps.Repos, deps.CacheHelper)

	return &Usecase{
		Configs:          config,
		Countries:        countries,
		FeatureCountries: featureCountries,
	}
}
