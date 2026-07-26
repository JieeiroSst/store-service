package application

import (
	"context"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adClickService struct {
	repo port.AdClickRepository
}

func NewAdClickService(repo port.AdClickRepository) port.AdClickUsecase {
	return &adClickService{repo: repo}
}

func (s *adClickService) Get(ctx context.Context, id uint) (*model.AdClick, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adClickService) List(ctx context.Context) ([]model.AdClick, error) {
	return s.repo.List(ctx)
}

func (s *adClickService) ListByAd(ctx context.Context, adID uint) ([]model.AdClick, error) {
	return s.repo.ListByAd(ctx, adID)
}
