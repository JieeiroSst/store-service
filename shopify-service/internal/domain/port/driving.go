package port

import (
	"context"
	"net/http"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
)

type ListInput struct {
	Limit  int
	Offset int
}

type CreateProductInput struct {
	Title           string
	DescriptionHTML string
	Vendor          string
	ProductType     string
	Tags            []string
	Status          model.ProductStatus
}

type ProductList struct {
	Items []model.Product
	Total int64
}

type OrderList struct {
	Items []model.Order
	Total int64
}

type SyncResult struct {
	Synced int `json:"synced"`
}

type ProductUsecase interface {
	CreateProduct(ctx context.Context, in CreateProductInput) (*model.Product, error)
	GetProduct(ctx context.Context, id int64) (*model.Product, error)
	ListProducts(ctx context.Context, in ListInput) (*ProductList, error)
	SyncProducts(ctx context.Context) (*SyncResult, error)
}

type OrderUsecase interface {
	GetOrder(ctx context.Context, id int64) (*model.Order, error)
	ListOrders(ctx context.Context, in ListInput) (*OrderList, error)
	SyncOrders(ctx context.Context) (*SyncResult, error)
}

type WebhookUsecase interface {
	HandleWebhook(ctx context.Context, payload []byte, headers http.Header) error
}
