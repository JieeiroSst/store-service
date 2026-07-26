package application

import (
	"context"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adPositionService struct {
	repo port.AdPositionRepository
}

func NewAdPositionService(repo port.AdPositionRepository) port.AdPositionUsecase {
	return &adPositionService{repo: repo}
}

func (s *adPositionService) Create(ctx context.Context, position *model.AdPosition) error {
	return s.repo.Create(ctx, position)
}

func (s *adPositionService) Get(ctx context.Context, id uint) (*model.AdPosition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adPositionService) Update(ctx context.Context, position *model.AdPosition) error {
	return s.repo.Update(ctx, position)
}

func (s *adPositionService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *adPositionService) List(ctx context.Context) ([]model.AdPosition, error) {
	return s.repo.List(ctx)
}
