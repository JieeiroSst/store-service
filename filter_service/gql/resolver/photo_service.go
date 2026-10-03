package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) PhotoQuery() generated.PhotoQueryResolver { return &photoQueryResolver{r} }

type photoQueryResolver struct{ *Resolver }

func (r *photoQueryResolver) Composition(ctx context.Context, obj *model.PhotoQuery, id string) (*model.PhotoComposition, error) {
	return r.Clients.PhotoService.Composition(ctx, id)
}
