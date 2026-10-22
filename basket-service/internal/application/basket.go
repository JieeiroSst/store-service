package application

import (
	"context"
	"time"

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
	basket.ID = 0
	basket.Status = model.BasketStatusOpen
	basket.DateCreated = time.Now().UTC().Format(time.RFC3339Nano)
	basket.DateSubmitted, basket.DatePlaced, basket.DateMerged = "", "", ""
	basket.Currency, basket.TotalInclTax, basket.PaymentReference, basket.FulfillmentID = "", 0, "", ""
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

func (s *basketService) ListBasketsByUser(ctx context.Context, userID int) ([]model.Basket, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *basketService) UpdateBasket(ctx context.Context, basket *model.Basket) (*model.Basket, error) {
	current, err := s.repo.GetByID(ctx, basket.ID)
	if err != nil {
		return nil, err
	}
	basket.Status = current.Status
	basket.DateCreated = current.DateCreated
	basket.DateSubmitted = current.DateSubmitted
	basket.DatePlaced = current.DatePlaced
	basket.DateMerged = current.DateMerged
	basket.Currency = current.Currency
	basket.TotalInclTax = current.TotalInclTax
	basket.PaymentReference = current.PaymentReference
	basket.FulfillmentID = current.FulfillmentID
	if err := s.repo.Update(ctx, basket); err != nil {
		return nil, err
	}
	return basket, nil
}

func (s *basketService) DeleteBasket(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
