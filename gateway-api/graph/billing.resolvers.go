package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) BillingCreateAddress(ctx context.Context, input model.BillingAddressInput) (*model.BillingAddress, error) {
	return restCall[*model.BillingAddress](ctx, r.Proxy, "billing-service", rest.Request{Method: "POST", Path: "/addresses", Body: input}, "data")
}

func (r *mutationResolver) BillingUpdateAddress(ctx context.Context, id int, input model.BillingAddressInput) (*model.BillingAddress, error) {
	return restCall[*model.BillingAddress](ctx, r.Proxy, "billing-service", rest.Request{Method: "PUT", Path: rest.P("/addresses/%v", id), Body: input}, "data")
}

func (r *mutationResolver) BillingDeleteAddress(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "billing-service", rest.Request{Method: "DELETE", Path: rest.P("/addresses/%v", id)}, "")
}

func (r *mutationResolver) BillingCreateCustomer(ctx context.Context, input model.BillingCustomerInput) (*model.BillingCustomer, error) {
	return restCall[*model.BillingCustomer](ctx, r.Proxy, "billing-service", rest.Request{Method: "POST", Path: "/customers", Body: input}, "data")
}

func (r *mutationResolver) BillingUpdateCustomer(ctx context.Context, id int, input model.BillingCustomerInput) (*model.BillingCustomer, error) {
	return restCall[*model.BillingCustomer](ctx, r.Proxy, "billing-service", rest.Request{Method: "PUT", Path: rest.P("/customers/%v", id), Body: input}, "data")
}

func (r *mutationResolver) BillingDeleteCustomer(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "billing-service", rest.Request{Method: "DELETE", Path: rest.P("/customers/%v", id)}, "")
}

func (r *mutationResolver) BillingCreatePlan(ctx context.Context, input model.BillingPlanInput) (*model.BillingPlan, error) {
	return restCall[*model.BillingPlan](ctx, r.Proxy, "billing-service", rest.Request{Method: "POST", Path: "/plans", Body: input}, "data")
}

func (r *mutationResolver) BillingUpdatePlan(ctx context.Context, id int, input model.BillingPlanInput) (*model.BillingPlan, error) {
	return restCall[*model.BillingPlan](ctx, r.Proxy, "billing-service", rest.Request{Method: "PUT", Path: rest.P("/plans/%v", id), Body: input}, "data")
}

func (r *mutationResolver) BillingDeletePlan(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "billing-service", rest.Request{Method: "DELETE", Path: rest.P("/plans/%v", id)}, "")
}

func (r *mutationResolver) BillingCreateSubscription(ctx context.Context, input model.BillingSubscriptionInput) (*model.BillingSubscription, error) {
	return restCall[*model.BillingSubscription](ctx, r.Proxy, "billing-service", rest.Request{Method: "POST", Path: "/subscriptions", Body: input}, "data")
}

func (r *mutationResolver) BillingUpdateSubscription(ctx context.Context, id int, input model.BillingSubscriptionInput) (*model.BillingSubscription, error) {
	return restCall[*model.BillingSubscription](ctx, r.Proxy, "billing-service", rest.Request{Method: "PUT", Path: rest.P("/subscriptions/%v", id), Body: input}, "data")
}

func (r *mutationResolver) BillingDeleteSubscription(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "billing-service", rest.Request{Method: "DELETE", Path: rest.P("/subscriptions/%v", id)}, "")
}

func (r *mutationResolver) BillingCreateInvoice(ctx context.Context, input model.BillingInvoiceInput) (*model.BillingInvoice, error) {
	return restCall[*model.BillingInvoice](ctx, r.Proxy, "billing-service", rest.Request{Method: "POST", Path: "/invoices", Body: input}, "data")
}

func (r *mutationResolver) BillingUpdateInvoice(ctx context.Context, id int, input model.BillingInvoiceInput) (*model.BillingInvoice, error) {
	return restCall[*model.BillingInvoice](ctx, r.Proxy, "billing-service", rest.Request{Method: "PUT", Path: rest.P("/invoices/%v", id), Body: input}, "data")
}

func (r *mutationResolver) BillingDeleteInvoice(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "billing-service", rest.Request{Method: "DELETE", Path: rest.P("/invoices/%v", id)}, "")
}

func (r *mutationResolver) BillingCreateTransaction(ctx context.Context, input model.BillingTransactionInput) (*model.BillingTransaction, error) {
	return restCall[*model.BillingTransaction](ctx, r.Proxy, "billing-service", rest.Request{Method: "POST", Path: "/transactions", Body: input}, "data")
}

func (r *mutationResolver) BillingUpdateTransaction(ctx context.Context, id int, input model.BillingTransactionInput) (*model.BillingTransaction, error) {
	return restCall[*model.BillingTransaction](ctx, r.Proxy, "billing-service", rest.Request{Method: "PUT", Path: rest.P("/transactions/%v", id), Body: input}, "data")
}

func (r *mutationResolver) BillingDeleteTransaction(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "billing-service", rest.Request{Method: "DELETE", Path: rest.P("/transactions/%v", id)}, "")
}

func (r *queryResolver) BillingAddresses(ctx context.Context) ([]*model.BillingAddress, error) {
	return restCall[[]*model.BillingAddress](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: "/addresses"}, "data")
}

func (r *queryResolver) BillingAddress(ctx context.Context, id int) (*model.BillingAddress, error) {
	return restCall[*model.BillingAddress](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: rest.P("/addresses/%v", id)}, "data")
}

func (r *queryResolver) BillingCustomers(ctx context.Context) ([]*model.BillingCustomer, error) {
	return restCall[[]*model.BillingCustomer](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: "/customers"}, "data")
}

func (r *queryResolver) BillingCustomer(ctx context.Context, id int) (*model.BillingCustomer, error) {
	return restCall[*model.BillingCustomer](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: rest.P("/customers/%v", id)}, "data")
}

func (r *queryResolver) BillingPlans(ctx context.Context) ([]*model.BillingPlan, error) {
	return restCall[[]*model.BillingPlan](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: "/plans"}, "data")
}

func (r *queryResolver) BillingPlan(ctx context.Context, id int) (*model.BillingPlan, error) {
	return restCall[*model.BillingPlan](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: rest.P("/plans/%v", id)}, "data")
}

func (r *queryResolver) BillingSubscriptions(ctx context.Context) ([]*model.BillingSubscription, error) {
	return restCall[[]*model.BillingSubscription](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: "/subscriptions"}, "data")
}

func (r *queryResolver) BillingSubscription(ctx context.Context, id int) (*model.BillingSubscription, error) {
	return restCall[*model.BillingSubscription](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: rest.P("/subscriptions/%v", id)}, "data")
}

func (r *queryResolver) BillingInvoices(ctx context.Context) ([]*model.BillingInvoice, error) {
	return restCall[[]*model.BillingInvoice](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: "/invoices"}, "data")
}

func (r *queryResolver) BillingInvoice(ctx context.Context, id int) (*model.BillingInvoice, error) {
	return restCall[*model.BillingInvoice](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: rest.P("/invoices/%v", id)}, "data")
}

func (r *queryResolver) BillingTransactions(ctx context.Context) ([]*model.BillingTransaction, error) {
	return restCall[[]*model.BillingTransaction](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: "/transactions"}, "data")
}

func (r *queryResolver) BillingTransaction(ctx context.Context, id int) (*model.BillingTransaction, error) {
	return restCall[*model.BillingTransaction](ctx, r.Proxy, "billing-service", rest.Request{Method: "GET", Path: rest.P("/transactions/%v", id)}, "data")
}
