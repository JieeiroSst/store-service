package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) IntPayQuery() generated.IntPayQueryResolver { return &intPayQueryResolver{r} }

type intPayQueryResolver struct{ *Resolver }

func (r *intPayQueryResolver) Payment(ctx context.Context, obj *model.IntPayQuery, id string) (*model.IntPayPayment, error) {
	return r.Clients.IntegratedPaymentService.Payment(ctx, id)
}

func (r *intPayQueryResolver) PaymentStatus(ctx context.Context, obj *model.IntPayQuery, id string) (*string, error) {
	return r.Clients.IntegratedPaymentService.PaymentStatus(ctx, id)
}
