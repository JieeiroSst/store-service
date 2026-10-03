package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ShopifyQuery() generated.ShopifyQueryResolver { return &shopifyQueryResolver{r} }

type shopifyQueryResolver struct{ *Resolver }

func (r *shopifyQueryResolver) Products(ctx context.Context, obj *model.ShopifyQuery, limit *int, offset *int) (*model.ShopifyProductList, error) {
	return r.Clients.ShopifyService.Products(ctx, limit, offset)
}

func (r *shopifyQueryResolver) Product(ctx context.Context, obj *model.ShopifyQuery, id int) (*model.ShopifyProduct, error) {
	return r.Clients.ShopifyService.Product(ctx, id)
}

func (r *shopifyQueryResolver) Orders(ctx context.Context, obj *model.ShopifyQuery, limit *int, offset *int) (*model.ShopifyOrderList, error) {
	return r.Clients.ShopifyService.Orders(ctx, limit, offset)
}

func (r *shopifyQueryResolver) Order(ctx context.Context, obj *model.ShopifyQuery, id int) (*model.ShopifyOrder, error) {
	return r.Clients.ShopifyService.Order(ctx, id)
}
