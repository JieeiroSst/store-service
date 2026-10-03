package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) OllamaPyQuery() generated.OllamaPyQueryResolver { return &ollamaPyQueryResolver{r} }

type ollamaPyQueryResolver struct{ *Resolver }

func (r *ollamaPyQueryResolver) Tags(ctx context.Context, obj *model.OllamaPyQuery) (*model.OllamaPyTags, error) {
	return r.Clients.OllamaModelPy.Tags(ctx)
}
