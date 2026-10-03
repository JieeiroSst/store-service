package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) PaymentQuery() generated.PaymentQueryResolver { return &paymentQueryResolver{r} }

type paymentQueryResolver struct{ *Resolver }

func (r *paymentQueryResolver) Payments(ctx context.Context, obj *model.PaymentQuery, provider *string, status *string, from *string, to *string, limit *int, offset *int) (*model.PaymentPaymentList, error) {
	return r.Clients.PaymentService.Payments(ctx, provider, status, from, to, limit, offset)
}

func (r *paymentQueryResolver) Payment(ctx context.Context, obj *model.PaymentQuery, id int) (*model.PaymentPayment, error) {
	return r.Clients.PaymentService.Payment(ctx, id)
}

func (r *paymentQueryResolver) Transactions(ctx context.Context, obj *model.PaymentQuery, id int) ([]*model.PaymentTransaction, error) {
	return r.Clients.PaymentService.Transactions(ctx, id)
}
