package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) IntPayCreatePayment(ctx context.Context, input model.IntPayCreatePaymentInput) (*model.IntPayPaymentResult, error) {
	return restCall[*model.IntPayPaymentResult](ctx, r.Proxy, "integrated-payment-service", rest.Request{Method: "POST", Path: "/api/v1/payments/", Body: input}, "data")
}

func (r *mutationResolver) IntPayProcessPayment(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "integrated-payment-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payments/%v/process", id), Body: map[string]any{}}, "data")
}

func (r *mutationResolver) IntPayRefundPayment(ctx context.Context, id string, amount float64) (any, error) {
	return restCall[any](ctx, r.Proxy, "integrated-payment-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payments/%v/refund", id), Body: map[string]any{"amount": amount}}, "data")
}

func (r *queryResolver) MyIntPayPayments(ctx context.Context) ([]*model.IntPayPayment, error) {
	return restCall[[]*model.IntPayPayment](ctx, r.Proxy, "integrated-payment-service", rest.Request{Method: "GET", Path: "/api/v1/me/payments"}, "data")
}

func (r *queryResolver) IntPayPayment(ctx context.Context, id string) (*model.IntPayPayment, error) {
	return restCall[*model.IntPayPayment](ctx, r.Proxy, "integrated-payment-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payments/%v", id)}, "data")
}

func (r *queryResolver) IntPayPaymentStatus(ctx context.Context, id string) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "integrated-payment-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payments/%v/status", id)}, "data")
}
