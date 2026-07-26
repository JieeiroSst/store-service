package application

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type basketService struct {
	repo port.BasketRepository
}

func NewBasketService(repo port.BasketRepository) port.BasketUsecase {
	return &basketService{repo: repo}
}

func (s *basketService) CreateBasket(ctx context.Context, basket *model.Basket) (*model.Basket, error) {
	if err := s.repo.Create(ctx, basket); err != nil {
		return nil, err
	}
	return basket, nil
}

func (s *basketService) GetBasket(ctx context.Context, id int) (*model.Basket, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *basketService) ListBaskets(ctx context.Context) ([]model.Basket, error) {
	return s.repo.List(ctx)
}

func (s *basketService) UpdateBasket(ctx context.Context, basket *model.Basket) (*model.Basket, error) {
	if _, err := s.repo.GetByID(ctx, basket.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, basket); err != nil {
		return nil, err
	}
	return basket, nil
}

func (s *basketService) DeleteBasket(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
