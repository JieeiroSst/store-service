package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RDeliveryRegister(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "delivery-service", rest.Request{Method: "POST", Path: "/api/v1/delivery/", Body: input}, "")
}

func (r *queryResolver) RDeliveryDeliveries(ctx context.Context, limit *int, page *int, sort *string) (*model.RDeliveryPage, error) {
	return restCall[*model.RDeliveryPage](ctx, r.Proxy, "delivery-service", rest.Request{Method: "GET", Path: "/api/v1/delivery/", Query: restQuery("Limit", limit, "Page", page, "Sort", sort)}, "data")
}
