package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) AutoPayCreateSubscription(ctx context.Context, input model.AutoPaySubscriptionInput) (*model.AutoPaySubscription, error) {
	return restCall[*model.AutoPaySubscription](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "POST", Path: "/subscriptions", Body: input}, "data")
}

func (r *mutationResolver) AutoPayCancelSubscription(ctx context.Context, id string) (*model.AutoPaySubscription, error) {
	return restCall[*model.AutoPaySubscription](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "POST", Path: rest.P("/subscriptions/%v/cancel", id), Body: map[string]any{}}, "data")
}

func (r *mutationResolver) AutoPayProcessRenewals(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "POST", Path: "/subscriptions/renewals/process", Body: map[string]any{}}, "data")
}

func (r *mutationResolver) AutoPayAddPaymentMethod(ctx context.Context, input model.AutoPayPaymentMethodInput) (*model.AutoPayPaymentMethod, error) {
	return restCall[*model.AutoPayPaymentMethod](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "POST", Path: "/payment-methods", Body: input}, "data")
}

func (r *mutationResolver) AutoPayDeletePaymentMethod(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "DELETE", Path: rest.P("/payment-methods/%v", id)}, "")
}

func (r *mutationResolver) AutoPaySetDefaultPaymentMethod(ctx context.Context, id string, userID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "POST", Path: rest.P("/payment-methods/%v/default", id), Body: map[string]any{"userId": userID}}, "data")
}

func (r *queryResolver) AutoPaySubscription(ctx context.Context, id string) (*model.AutoPaySubscription, error) {
	return restCall[*model.AutoPaySubscription](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "GET", Path: rest.P("/subscriptions/%v", id)}, "data")
}

func (r *queryResolver) AutoPaySubscriptionTransactions(ctx context.Context, id string) ([]*model.AutoPayTransaction, error) {
	return restCall[[]*model.AutoPayTransaction](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "GET", Path: rest.P("/subscriptions/%v/transactions", id)}, "data")
}

func (r *queryResolver) AutoPaySubscriptionInvoices(ctx context.Context, id string) ([]*model.AutoPayInvoice, error) {
	return restCall[[]*model.AutoPayInvoice](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "GET", Path: rest.P("/subscriptions/%v/invoices", id)}, "data")
}

func (r *queryResolver) AutoPayPaymentMethods(ctx context.Context, userID string) ([]*model.AutoPayPaymentMethod, error) {
	return restCall[[]*model.AutoPayPaymentMethod](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "GET", Path: "/payment-methods", Query: restQuery("userId", userID)}, "data")
}

func (r *queryResolver) AutoPayInvoices(ctx context.Context, userID string) ([]*model.AutoPayInvoice, error) {
	return restCall[[]*model.AutoPayInvoice](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "GET", Path: "/invoices", Query: restQuery("userId", userID)}, "data")
}

func (r *queryResolver) AutoPayInvoice(ctx context.Context, id string) (*model.AutoPayInvoice, error) {
	return restCall[*model.AutoPayInvoice](ctx, r.Proxy, "automatic-payment-service", rest.Request{Method: "GET", Path: rest.P("/invoices/%v", id)}, "data")
}
