package application

import (
	"context"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adCategoryService struct {
	repo port.AdCategoryRepository
}

func NewAdCategoryService(repo port.AdCategoryRepository) port.AdCategoryUsecase {
	return &adCategoryService{repo: repo}
}

func (s *adCategoryService) Create(ctx context.Context, category *model.AdCategory) error {
	return s.repo.Create(ctx, category)
}

func (s *adCategoryService) Get(ctx context.Context, id uint) (*model.AdCategory, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adCategoryService) Update(ctx context.Context, category *model.AdCategory) error {
	return s.repo.Update(ctx, category)
}

func (s *adCategoryService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *adCategoryService) List(ctx context.Context) ([]model.AdCategory, error) {
	return s.repo.List(ctx)
}
