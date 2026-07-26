package application

import (
	"context"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adPositionMappingService struct {
	repo port.AdPositionMappingRepository
}

func NewAdPositionMappingService(repo port.AdPositionMappingRepository) port.AdPositionMappingUsecase {
	return &adPositionMappingService{repo: repo}
}

func (s *adPositionMappingService) Create(ctx context.Context, mapping *model.AdPositionMapping) error {
	return s.repo.Create(ctx, mapping)
}

func (s *adPositionMappingService) Get(ctx context.Context, id uint) (*model.AdPositionMapping, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adPositionMappingService) Update(ctx context.Context, mapping *model.AdPositionMapping) error {
	return s.repo.Update(ctx, mapping)
}

func (s *adPositionMappingService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *adPositionMappingService) List(ctx context.Context) ([]model.AdPositionMapping, error) {
	return s.repo.List(ctx)
}
