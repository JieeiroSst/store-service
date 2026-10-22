package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type basketLineService struct {
	repo      port.BasketLineRepository
	baskets   port.BasketRepository
	catalogue port.Catalogue
}

func NewBasketLineService(repo port.BasketLineRepository, baskets port.BasketRepository, catalogue port.Catalogue) port.BasketLineUsecase {
	return &basketLineService{repo: repo, baskets: baskets, catalogue: catalogue}
}

func (s *basketLineService) price(ctx context.Context, line *model.BasketLine) error {
	if line.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", common.ErrInvalidRequest)
	}
	if line.StockrecordID <= 0 {
		return fmt.Errorf("%w: stockrecord_id is required", common.ErrInvalidRequest)
	}
	record, err := s.catalogue.StockRecord(ctx, line.StockrecordID)
	if err != nil {
		return err
	}
	if !record.CanSupply(line.Quantity) {
		return common.ErrOutOfStock
	}
	line.ProductID = record.ProductID
	line.PriceCurrency = record.PriceCurrency
	line.PriceExclTax = record.Price
	line.PriceInclTax = record.Price
	return nil
}

func (s *basketLineService) openBasket(ctx context.Context, id int) error {
	basket, err := s.baskets.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !basket.IsOpen() {
		return fmt.Errorf("%w: basket %d is already %s", common.ErrConflict, id, basket.Status)
	}
	return nil
}

func (s *basketLineService) CreateBasketLine(ctx context.Context, line *model.BasketLine) (*model.BasketLine, error) {
	if err := s.openBasket(ctx, line.BasketID); err != nil {
		return nil, err
	}
	existing, err := s.repo.ListByBasket(ctx, line.BasketID)
	if err != nil {
		return nil, err
	}
	for _, e := range existing {
		if e.StockrecordID == line.StockrecordID && line.StockrecordID > 0 {
			merged := e
			merged.Quantity += line.Quantity
			if err := s.price(ctx, &merged); err != nil {
				return nil, err
			}
			if err := s.repo.Update(ctx, &merged); err != nil {
				return nil, err
			}
			return &merged, nil
		}
	}
	line.ID = 0
	if line.LineReference == "" {
		line.LineReference = fmt.Sprintf("%d_%d", line.BasketID, line.StockrecordID)
	}
	if err := s.price(ctx, line); err != nil {
		return nil, err
	}
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

func (s *basketLineService) ListBasketLinesByBasket(ctx context.Context, basketID int) ([]model.BasketLine, error) {
	return s.repo.ListByBasket(ctx, basketID)
}

func (s *basketLineService) UpdateBasketLine(ctx context.Context, line *model.BasketLine) (*model.BasketLine, error) {
	current, err := s.repo.GetByID(ctx, line.ID)
	if err != nil {
		return nil, err
	}
	if err := s.openBasket(ctx, current.BasketID); err != nil {
		return nil, err
	}
	updated := *current
	updated.Quantity = line.Quantity
	if line.LineReference != "" {
		updated.LineReference = line.LineReference
	}
	if err := s.price(ctx, &updated); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (s *basketLineService) DeleteBasketLine(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
