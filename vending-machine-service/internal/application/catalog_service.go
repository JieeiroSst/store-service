package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/google/uuid"
)

type catalogService struct {
	categories port.CategoryRepository
	products   port.ProductRepository
	now        func() time.Time
}

func NewCatalogService(categories port.CategoryRepository, products port.ProductRepository) port.CatalogService {
	return &catalogService{categories: categories, products: products, now: time.Now}
}

func (s *catalogService) CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	now := s.now().UTC()
	c.ID = uuid.NewString()
	c.CreatedAt, c.UpdatedAt = now, now
	if err := s.categories.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *catalogService) ListCategories(ctx context.Context, page domain.PageRequest) (domain.Page[domain.Category], error) {
	return s.categories.List(ctx, page)
}

func (s *catalogService) CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	now := s.now().UTC()
	p.ID = uuid.NewString()
	p.CreatedAt, p.UpdatedAt = now, now
	if err := s.products.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *catalogService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.products.Get(ctx, id)
}

func (s *catalogService) ListProducts(ctx context.Context, categoryID string, page domain.PageRequest) (domain.Page[domain.Product], error) {
	return s.products.List(ctx, categoryID, page)
}

func (s *catalogService) UpdateProduct(ctx context.Context, id string, patch domain.ProductPatch) (*domain.Product, error) {
	p, err := s.products.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := p.Apply(patch); err != nil {
		return nil, err
	}
	p.UpdatedAt = s.now().UTC()
	if err := s.products.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
