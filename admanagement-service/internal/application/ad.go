package application

import (
	"context"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adService struct {
	repo port.AdRepository
}

func NewAdService(repo port.AdRepository) port.AdUsecase {
	return &adService{repo: repo}
}

func (s *adService) Create(ctx context.Context, ad *model.Ad) error {
	return s.repo.Create(ctx, ad)
}

func (s *adService) Get(ctx context.Context, id uint) (*model.Ad, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adService) Update(ctx context.Context, ad *model.Ad) error {
	return s.repo.Update(ctx, ad)
}

func (s *adService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *adService) List(ctx context.Context) ([]model.Ad, error) {
	return s.repo.List(ctx)
}

func (s *adService) ListByCampaign(ctx context.Context, campaignID uint) ([]model.Ad, error) {
	return s.repo.ListByCampaign(ctx, campaignID)
}
