package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type stockRecordService struct {
	repo     port.StockRecordRepository
	products port.ProductRepository
}

func NewStockRecordService(repo port.StockRecordRepository, products port.ProductRepository) port.StockRecordService {
	return &stockRecordService{repo: repo, products: products}
}

func prepareStock(s *model.StockRecord) error {
	s.PartnerSKU = strings.TrimSpace(s.PartnerSKU)
	s.PriceCurrency = strings.ToUpper(strings.TrimSpace(s.PriceCurrency))
	if s.PriceCurrency == "" {
		s.PriceCurrency = "VND"
	}
	if len(s.PriceCurrency) != 3 {
		return model.Invalid("price_currency must be a 3-letter ISO code")
	}
	if s.Price < 0 {
		return model.Invalid("price must not be negative")
	}
	if s.NumInStock != nil && *s.NumInStock < 0 {
		return model.Invalid("num_in_stock must not be negative")
	}
	if s.NumAllocated < 0 {
		return model.Invalid("num_allocated must not be negative")
	}
	return nil
}

func (s *stockRecordService) Create(ctx context.Context, r model.StockRecord) (*model.StockRecord, error) {
	if err := prepareStock(&r); err != nil {
		return nil, err
	}
	if _, err := s.products.Get(ctx, r.ProductID); err != nil {
		return nil, err
	}
	r.ID = 0
	r.NumAllocated = 0
	if err := s.repo.Create(ctx, &r); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, r.ID)
}

func (s *stockRecordService) Get(ctx context.Context, id int64) (*model.StockRecord, error) {
	return s.repo.Get(ctx, id)
}

func (s *stockRecordService) ListByProduct(ctx context.Context, productID int64) ([]model.StockRecord, error) {
	if _, err := s.products.Get(ctx, productID); err != nil {
		return nil, err
	}
	return s.repo.ListByProducts(ctx, []int64{productID})
}

func (s *stockRecordService) Update(ctx context.Context, r model.StockRecord) (*model.StockRecord, error) {
	current, err := s.repo.Get(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	r.ProductID = current.ProductID
	r.NumAllocated = current.NumAllocated
	if err := prepareStock(&r); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, &r); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, r.ID)
}

func (s *stockRecordService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *stockRecordService) Allocate(ctx context.Context, id int64, qty int) (*model.StockRecord, error) {
	if qty <= 0 {
		return nil, model.Invalid("quantity must be positive")
	}
	if err := s.repo.Allocate(ctx, id, qty); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

func (s *stockRecordService) Release(ctx context.Context, id int64, qty int) (*model.StockRecord, error) {
	if qty <= 0 {
		return nil, model.Invalid("quantity must be positive")
	}
	if err := s.repo.Release(ctx, id, qty); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}
