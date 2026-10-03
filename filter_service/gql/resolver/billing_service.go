package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) BillingQuery() generated.BillingQueryResolver { return &billingQueryResolver{r} }

type billingQueryResolver struct{ *Resolver }

func (r *billingQueryResolver) Addresses(ctx context.Context, obj *model.BillingQuery) ([]*model.BillingAddress, error) {
	return r.Clients.BillingService.Addresses(ctx)
}

func (r *billingQueryResolver) Address(ctx context.Context, obj *model.BillingQuery, id int) (*model.BillingAddress, error) {
	return r.Clients.BillingService.Address(ctx, id)
}

func (r *billingQueryResolver) Customers(ctx context.Context, obj *model.BillingQuery) ([]*model.BillingCustomer, error) {
	return r.Clients.BillingService.Customers(ctx)
}

func (r *billingQueryResolver) Customer(ctx context.Context, obj *model.BillingQuery, id int) (*model.BillingCustomer, error) {
	return r.Clients.BillingService.Customer(ctx, id)
}

func (r *billingQueryResolver) Plans(ctx context.Context, obj *model.BillingQuery) ([]*model.BillingPlan, error) {
	return r.Clients.BillingService.Plans(ctx)
}

func (r *billingQueryResolver) Plan(ctx context.Context, obj *model.BillingQuery, id int) (*model.BillingPlan, error) {
	return r.Clients.BillingService.Plan(ctx, id)
}

func (r *billingQueryResolver) Subscriptions(ctx context.Context, obj *model.BillingQuery) ([]*model.BillingSubscription, error) {
	return r.Clients.BillingService.Subscriptions(ctx)
}

func (r *billingQueryResolver) Subscription(ctx context.Context, obj *model.BillingQuery, id int) (*model.BillingSubscription, error) {
	return r.Clients.BillingService.Subscription(ctx, id)
}

func (r *billingQueryResolver) Invoices(ctx context.Context, obj *model.BillingQuery) ([]*model.BillingInvoice, error) {
	return r.Clients.BillingService.Invoices(ctx)
}

func (r *billingQueryResolver) Invoice(ctx context.Context, obj *model.BillingQuery, id int) (*model.BillingInvoice, error) {
	return r.Clients.BillingService.Invoice(ctx, id)
}

func (r *billingQueryResolver) Transactions(ctx context.Context, obj *model.BillingQuery) ([]*model.BillingTransaction, error) {
	return r.Clients.BillingService.Transactions(ctx)
}

func (r *billingQueryResolver) Transaction(ctx context.Context, obj *model.BillingQuery, id int) (*model.BillingTransaction, error) {
	return r.Clients.BillingService.Transaction(ctx, id)
}
