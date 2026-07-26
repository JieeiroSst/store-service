package application

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type basketLineService struct {
	repo port.BasketLineRepository
}

func NewBasketLineService(repo port.BasketLineRepository) port.BasketLineUsecase {
	return &basketLineService{repo: repo}
}

func (s *basketLineService) CreateBasketLine(ctx context.Context, line *model.BasketLine) (*model.BasketLine, error) {
	if err := s.repo.Create(ctx, line); err != nil {
		return nil, err
	}
	return line, nil
}

func (s *basketLineService) GetBasketLine(ctx context.Context, id int) (*model.BasketLine, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *basketLineService) ListBasketLines(ctx context.Context) ([]model.BasketLine, error) {
	return s.repo.List(ctx)
}

func (s *basketLineService) UpdateBasketLine(ctx context.Context, line *model.BasketLine) (*model.BasketLine, error) {
	if _, err := s.repo.GetByID(ctx, line.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, line); err != nil {
		return nil, err
	}
	return line, nil
}

func (s *basketLineService) DeleteBasketLine(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
