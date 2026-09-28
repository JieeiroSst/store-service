package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
)

type productService struct {
	repo    port.ProductRepository
	shopify port.ShopifyClient
}

func NewProductService(repo port.ProductRepository, shopify port.ShopifyClient) port.ProductUsecase {
	return &productService{repo: repo, shopify: shopify}
}

func (s *productService) CreateProduct(ctx context.Context, in port.CreateProductInput) (*model.Product, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return nil, fmt.Errorf("%w: title is required", port.ErrInvalidInput)
	}
	if in.Status != "" && !in.Status.Valid() {
		return nil, fmt.Errorf("%w: status must be ACTIVE, DRAFT or ARCHIVED", port.ErrInvalidInput)
	}

	created, err := s.shopify.CreateProduct(ctx, in)
	if err != nil {
		return nil, err
	}
	return s.repo.Upsert(ctx, created)
}

func (s *productService) GetProduct(ctx context.Context, id int64) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) ListProducts(ctx context.Context, in port.ListInput) (*port.ProductList, error) {
	limit, offset := normalizeList(in)
	items, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return &port.ProductList{Items: items, Total: total}, nil
}

func (s *productService) SyncProducts(ctx context.Context) (*port.SyncResult, error) {
	result := &port.SyncResult{}
	cursor := ""
	for {
		page, err := s.shopify.ListProducts(ctx, cursor)
		if err != nil {
			return result, err
		}
		for i := range page.Items {
			if _, err := s.repo.Upsert(ctx, &page.Items[i]); err != nil {
				return result, fmt.Errorf("upsert product %s: %w", page.Items[i].ShopifyID, err)
			}
			result.Synced++
		}
		if !page.HasNextPage {
			return result, nil
		}
		cursor = page.EndCursor
	}
}
