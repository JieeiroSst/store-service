package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ShopifyCreateProduct(ctx context.Context, title string, descriptionHTML *string, vendor *string, productType *string, tags []string, status *string) (*model.ShopifyProduct, error) {
	return restCall[*model.ShopifyProduct](ctx, r.Proxy, "shopify-service", rest.Request{Method: "POST", Path: "/api/v1/products", Body: map[string]any{"title": title, "description_html": descriptionHTML, "vendor": vendor, "product_type": productType, "tags": tags, "status": status}}, "")
}

func (r *mutationResolver) ShopifySyncProducts(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "shopify-service", rest.Request{Method: "POST", Path: "/api/v1/products/sync", Body: map[string]any{}}, "")
}

func (r *mutationResolver) ShopifySyncOrders(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "shopify-service", rest.Request{Method: "POST", Path: "/api/v1/orders/sync", Body: map[string]any{}}, "")
}

func (r *queryResolver) ShopifyProducts(ctx context.Context, limit *int, offset *int) (*model.ShopifyProductList, error) {
	return restCall[*model.ShopifyProductList](ctx, r.Proxy, "shopify-service", rest.Request{Method: "GET", Path: "/api/v1/products", Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) ShopifyProduct(ctx context.Context, id int) (*model.ShopifyProduct, error) {
	return restCall[*model.ShopifyProduct](ctx, r.Proxy, "shopify-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/products/%v", id)}, "")
}

func (r *queryResolver) ShopifyOrders(ctx context.Context, limit *int, offset *int) (*model.ShopifyOrderList, error) {
	return restCall[*model.ShopifyOrderList](ctx, r.Proxy, "shopify-service", rest.Request{Method: "GET", Path: "/api/v1/orders", Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) ShopifyOrder(ctx context.Context, id int) (*model.ShopifyOrder, error) {
	return restCall[*model.ShopifyOrder](ctx, r.Proxy, "shopify-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "")
}
