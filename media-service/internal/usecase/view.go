package usecase

import (
	"context"

	"github.com/JIeeiroSst/media-service/dto"
	"github.com/JIeeiroSst/media-service/internal/repository"
	"github.com/JIeeiroSst/media-service/internal/usecase/build"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type View interface {
	SaveView(ctx context.Context, view dto.View) error
	FindByID(ctx context.Context, viewID int) (*dto.View, error)
}

type ViewUsecase struct {
	repo  *repository.Repository
	cache expire.CacheHelper
}

func NewViewUsecase(repo *repository.Repository,
	cache expire.CacheHelper) *ViewUsecase {
	return &ViewUsecase{
		repo:  repo,
		cache: cache,
	}
}

func (u *ViewUsecase) SaveView(ctx context.Context, view dto.View) error {
	model := build.BuildSaveView(view)

	if err := u.repo.View.SaveView(ctx, model); err != nil {
		return err
	}

	return nil
}

func (u *ViewUsecase) FindByID(ctx context.Context, viewID int) (*dto.View, error) {
	view, err := u.repo.View.FindByID(ctx, viewID)
	if err != nil {
		return nil, err
	}

	return build.BuildView(view), nil
}
