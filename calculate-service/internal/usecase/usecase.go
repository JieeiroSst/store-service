package usecase

import (
	"github.com/JIeeiroSst/calculate-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Usecase struct {
	Campaigns
	Users
}

type Dependency struct {
	Repos       *repository.Repository
	CacheHelper expire.CacheHelper
}

func NewUsecase(deps Dependency) *Usecase {
	campaigns := NewCampaignUsecase(deps.Repos, deps.CacheHelper)
	users := NewUserUsecase(deps.Repos)

	return &Usecase{
		Campaigns: campaigns,
		Users:     users,
	}
}
