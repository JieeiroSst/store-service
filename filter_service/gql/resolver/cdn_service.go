package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) CdnQuery() generated.CdnQueryResolver { return &cdnQueryResolver{r} }

type cdnQueryResolver struct{ *Resolver }

func (r *cdnQueryResolver) Files(ctx context.Context, obj *model.CdnQuery, limit *int, offset *int) (*model.CdnFileList, error) {
	return r.Clients.CdnService.Files(ctx, limit, offset)
}

func (r *cdnQueryResolver) File(ctx context.Context, obj *model.CdnQuery, id string) (*model.CdnFile, error) {
	return r.Clients.CdnService.File(ctx, id)
}
