package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) DrawImageQuery() generated.DrawImageQueryResolver {
	return &drawImageQueryResolver{r}
}

type drawImageQueryResolver struct{ *Resolver }

func (r *drawImageQueryResolver) ImageInfo(ctx context.Context, obj *model.DrawImageQuery, id string) (*model.DrawImageCollage, error) {
	return r.Clients.DrawImageService.ImageInfo(ctx, id)
}
