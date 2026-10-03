package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) SardQuery() generated.SardQueryResolver { return &sardQueryResolver{r} }

type sardQueryResolver struct{ *Resolver }

func (r *sardQueryResolver) Info(ctx context.Context, obj *model.SardQuery) (*model.SardInfo, error) {
	return r.Clients.SardService.Info(ctx)
}
