package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) EkycQuery() generated.EkycQueryResolver { return &ekycQueryResolver{r} }

type ekycQueryResolver struct{ *Resolver }

func (r *ekycQueryResolver) Status(ctx context.Context, obj *model.EkycQuery, userID string) (*model.EkycEkycStatus, error) {
	return r.Clients.EkycService.Status(ctx, userID)
}
