package port

import (
	"context"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
)

type ProductRepository interface {
	Upsert(ctx context.Context, product *model.Product) (*model.Product, error)
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	List(ctx context.Context, limit, offset int) ([]model.Product, int64, error)
	DeleteByShopifyID(ctx context.Context, shopifyID string) error
}

type OrderRepository interface {
	Upsert(ctx context.Context, order *model.Order) (*model.Order, error)
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	List(ctx context.Context, limit, offset int) ([]model.Order, int64, error)
}

type WebhookEventRepository interface {
	Exists(ctx context.Context, webhookID string) (bool, error)
	Create(ctx context.Context, event *model.WebhookEvent) error
}

type Page[T any] struct {
	Items       []T
	HasNextPage bool
	EndCursor   string
}

type ShopifyClient interface {
	ListProducts(ctx context.Context, after string) (*Page[model.Product], error)
	GetProduct(ctx context.Context, shopifyID string) (*model.Product, error)
	CreateProduct(ctx context.Context, in CreateProductInput) (*model.Product, error)
	ListOrders(ctx context.Context, after string) (*Page[model.Order], error)
	GetOrder(ctx context.Context, shopifyID string) (*model.Order, error)
}

type WebhookVerifier interface {
	Verify(payload []byte, signature string) bool
}
