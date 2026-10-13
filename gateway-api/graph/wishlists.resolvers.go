package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) WishlistCreateCustomer(ctx context.Context, name string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "POST", Path: rest.P("/customer/%v", name), Body: map[string]any{}}, "data")
}

func (r *queryResolver) WishlistGreet(ctx context.Context) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: "/greet"}, "data")
}

func (r *queryResolver) WishlistCustomers(ctx context.Context) ([]*model.WishlistCustomer, error) {
	return restCall[[]*model.WishlistCustomer](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: "/customer"}, "data")
}
