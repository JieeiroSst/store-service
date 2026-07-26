package application

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type basketLineAttributeService struct {
	repo port.BasketLineAttributeRepository
}

func NewBasketLineAttributeService(repo port.BasketLineAttributeRepository) port.BasketLineAttributeUsecase {
	return &basketLineAttributeService{repo: repo}
}

func (s *basketLineAttributeService) CreateBasketLineAttribute(ctx context.Context, attribute *model.BasketLineAttribute) (*model.BasketLineAttribute, error) {
	if err := s.repo.Create(ctx, attribute); err != nil {
		return nil, err
	}
	return attribute, nil
}

func (s *basketLineAttributeService) GetBasketLineAttribute(ctx context.Context, id int) (*model.BasketLineAttribute, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *basketLineAttributeService) ListBasketLineAttributes(ctx context.Context) ([]model.BasketLineAttribute, error) {
	return s.repo.List(ctx)
}

func (s *basketLineAttributeService) UpdateBasketLineAttribute(ctx context.Context, attribute *model.BasketLineAttribute) (*model.BasketLineAttribute, error) {
	if _, err := s.repo.GetByID(ctx, attribute.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, attribute); err != nil {
		return nil, err
	}
	return attribute, nil
}

func (s *basketLineAttributeService) DeleteBasketLineAttribute(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
