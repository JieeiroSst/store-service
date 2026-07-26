package application

import (
	"context"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adImpressionService struct {
	repo port.AdImpressionRepository
}

func NewAdImpressionService(repo port.AdImpressionRepository) port.AdImpressionUsecase {
	return &adImpressionService{repo: repo}
}

func (s *adImpressionService) Get(ctx context.Context, id uint) (*model.AdImpression, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adImpressionService) List(ctx context.Context) ([]model.AdImpression, error) {
	return s.repo.List(ctx)
}

func (s *adImpressionService) ListByAd(ctx context.Context, adID uint) ([]model.AdImpression, error) {
	return s.repo.ListByAd(ctx, adID)
}
