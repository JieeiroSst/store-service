package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RConsumerQuery() generated.RConsumerQueryResolver {
	return &rConsumerQueryResolver{r}
}

type rConsumerQueryResolver struct{ *Resolver }

func (r *rConsumerQueryResolver) Consumers(ctx context.Context, obj *model.RConsumerQuery, limit *int, page *int, sort *string) (*model.RConsumerPage, error) {
	return r.Clients.RestaurantConsumerService.Consumers(ctx, limit, page, sort)
}
