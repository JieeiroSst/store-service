package usecase

import (
	"context"

	"github.com/JIeeiroSst/media-service/dto"
	"github.com/JIeeiroSst/media-service/internal/repository"
	"github.com/JIeeiroSst/media-service/internal/usecase/build"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Subscription interface {
	SaveSubscription(ctx context.Context, subscription dto.Subscription) error
}

type SubscriptionUsecase struct {
	repo  *repository.Repository
	cache expire.CacheHelper
}

func NewSubscriptionUsecase(repo *repository.Repository,
	cache expire.CacheHelper) *SubscriptionUsecase {
	return &SubscriptionUsecase{
		repo:  repo,
		cache: cache,
	}
}

func (u *SubscriptionUsecase) SaveSubscription(ctx context.Context, subscription dto.Subscription) error {
	model := build.BuildSubscription(subscription)

	if err := u.repo.Subscription.SaveSubscription(ctx, model); err != nil {
		return err
	}
	return nil
}
