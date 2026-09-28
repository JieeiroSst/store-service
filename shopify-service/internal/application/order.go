package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
)

type orderService struct {
	repo    port.OrderRepository
	shopify port.ShopifyClient
}

func NewOrderService(repo port.OrderRepository, shopify port.ShopifyClient) port.OrderUsecase {
	return &orderService{repo: repo, shopify: shopify}
}

func (s *orderService) GetOrder(ctx context.Context, id int64) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) ListOrders(ctx context.Context, in port.ListInput) (*port.OrderList, error) {
	limit, offset := normalizeList(in)
	items, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return &port.OrderList{Items: items, Total: total}, nil
}

func (s *orderService) SyncOrders(ctx context.Context) (*port.SyncResult, error) {
	result := &port.SyncResult{}
	cursor := ""
	for {
		page, err := s.shopify.ListOrders(ctx, cursor)
		if err != nil {
			return result, err
		}
		for i := range page.Items {
			if _, err := s.repo.Upsert(ctx, &page.Items[i]); err != nil {
				return result, fmt.Errorf("upsert order %s: %w", page.Items[i].ShopifyID, err)
			}
			result.Synced++
		}
		if !page.HasNextPage {
			return result, nil
		}
		cursor = page.EndCursor
	}
}
