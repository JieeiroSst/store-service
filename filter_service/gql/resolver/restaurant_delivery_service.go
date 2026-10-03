package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RDeliveryQuery() generated.RDeliveryQueryResolver {
	return &rDeliveryQueryResolver{r}
}

type rDeliveryQueryResolver struct{ *Resolver }

func (r *rDeliveryQueryResolver) Deliveries(ctx context.Context, obj *model.RDeliveryQuery, limit *int, page *int, sort *string) (*model.RDeliveryPage, error) {
	return r.Clients.RestaurantDeliveryService.Deliveries(ctx, limit, page, sort)
}
