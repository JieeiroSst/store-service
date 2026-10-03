package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) AutoPayQuery() generated.AutoPayQueryResolver { return &autoPayQueryResolver{r} }

type autoPayQueryResolver struct{ *Resolver }

func (r *autoPayQueryResolver) Subscription(ctx context.Context, obj *model.AutoPayQuery, id string) (*model.AutoPaySubscription, error) {
	return r.Clients.AutomaticPaymentService.Subscription(ctx, id)
}

func (r *autoPayQueryResolver) SubscriptionTransactions(ctx context.Context, obj *model.AutoPayQuery, id string) ([]*model.AutoPayTransaction, error) {
	return r.Clients.AutomaticPaymentService.SubscriptionTransactions(ctx, id)
}

func (r *autoPayQueryResolver) SubscriptionInvoices(ctx context.Context, obj *model.AutoPayQuery, id string) ([]*model.AutoPayInvoice, error) {
	return r.Clients.AutomaticPaymentService.SubscriptionInvoices(ctx, id)
}

func (r *autoPayQueryResolver) PaymentMethods(ctx context.Context, obj *model.AutoPayQuery, userID string) ([]*model.AutoPayPaymentMethod, error) {
	return r.Clients.AutomaticPaymentService.PaymentMethods(ctx, userID)
}

func (r *autoPayQueryResolver) Invoices(ctx context.Context, obj *model.AutoPayQuery, userID string) ([]*model.AutoPayInvoice, error) {
	return r.Clients.AutomaticPaymentService.Invoices(ctx, userID)
}

func (r *autoPayQueryResolver) Invoice(ctx context.Context, obj *model.AutoPayQuery, id string) (*model.AutoPayInvoice, error) {
	return r.Clients.AutomaticPaymentService.Invoice(ctx, id)
}
