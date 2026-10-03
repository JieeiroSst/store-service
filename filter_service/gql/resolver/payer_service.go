package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) PayerQuery() generated.PayerQueryResolver { return &payerQueryResolver{r} }

type payerQueryResolver struct{ *Resolver }

func (r *payerQueryResolver) Transaction(ctx context.Context, obj *model.PayerQuery, transactionID int) (*model.PayerTransactionResponse, error) {
	return r.Clients.PayerService.Transaction(ctx, transactionID)
}
