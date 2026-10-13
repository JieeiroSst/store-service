package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) PaymentCreatePayment(ctx context.Context, input model.PaymentCreateInput) (*model.PaymentPayment, error) {
	return restCall[*model.PaymentPayment](ctx, r.Proxy, "payment-service", rest.Request{Method: "POST", Path: "/api/v1/payments", Body: input}, "")
}

func (r *mutationResolver) PaymentRefundPayment(ctx context.Context, id string, amount *int) (*model.PaymentPayment, error) {
	return restCall[*model.PaymentPayment](ctx, r.Proxy, "payment-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payments/%v/refund", id), Body: map[string]any{"amount": amount}}, "")
}

func (r *queryResolver) Payments(ctx context.Context, provider *string, status *string, from *string, to *string, limit *int, offset *int) (*model.PaymentPaymentList, error) {
	return restCall[*model.PaymentPaymentList](ctx, r.Proxy, "payment-service", rest.Request{Method: "GET", Path: "/api/v1/payments", Query: restQuery("provider", provider, "status", status, "from", from, "to", to, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) Payment(ctx context.Context, id int) (*model.PaymentPayment, error) {
	return restCall[*model.PaymentPayment](ctx, r.Proxy, "payment-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payments/%v", id)}, "")
}

func (r *queryResolver) PaymentTransactions(ctx context.Context, id int) ([]*model.PaymentTransaction, error) {
	return restCall[[]*model.PaymentTransaction](ctx, r.Proxy, "payment-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payments/%v/transactions", id)}, "")
}
