package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RConsumerPlaceOrder(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "consumer-service", rest.Request{Method: "POST", Path: "/api/v1/consumer", Body: input}, "")
}

func (r *queryResolver) RConsumerConsumers(ctx context.Context, limit *int, page *int, sort *string) (*model.RConsumerPage, error) {
	return restCall[*model.RConsumerPage](ctx, r.Proxy, "consumer-service", rest.Request{Method: "GET", Path: "/api/v1/consumer", Query: restQuery("Limit", limit, "Page", page, "Sort", sort)}, "data")
}
