package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RAccountingAuthCart(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "accounting-service", rest.Request{Method: "POST", Path: "/api/v1/accounting", Body: input}, "")
}

func (r *mutationResolver) RAccountingMarkPaid(ctx context.Context, orderID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "accounting-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/payments/%v/pay", orderID), Body: map[string]any{}}, "")
}

func (r *queryResolver) RAccountingPayment(ctx context.Context, orderID int) (*model.RAccountingPayment, error) {
	return restCall[*model.RAccountingPayment](ctx, r.Proxy, "accounting-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payments/%v", orderID)}, "data")
}
