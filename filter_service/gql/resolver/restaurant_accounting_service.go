package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RAccountingQuery() generated.RAccountingQueryResolver {
	return &rAccountingQueryResolver{r}
}

type rAccountingQueryResolver struct{ *Resolver }

func (r *rAccountingQueryResolver) Payment(ctx context.Context, obj *model.RAccountingQuery, orderID int) (*model.RAccountingPayment, error) {
	return r.Clients.RestaurantAccountingService.Payment(ctx, orderID)
}
