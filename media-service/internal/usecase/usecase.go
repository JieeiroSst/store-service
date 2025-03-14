package usecase

import (
	"github.com/JIeeiroSst/media-service/internal/proxy/cloudflare"
	"github.com/JIeeiroSst/media-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Usecase struct {
	Videos
	Subscription
	View
}

type Dependency struct {
	Repos       *repository.Repository
	CacheHelper expire.CacheHelper
	Cloudflare  cloudflare.CloudflareProxy
}

func NewUsecase(deps Dependency) *Usecase {
	subscription := NewSubscriptionUsecase(deps.Repos, deps.CacheHelper)
	view := NewViewUsecase(deps.Repos, deps.CacheHelper)
	videos := NewVideoUsecase(deps.Repos, deps.CacheHelper, deps.Cloudflare, view)
	return &Usecase{
		Videos:       videos,
		Subscription: subscription,
		View:         view,
	}
}
