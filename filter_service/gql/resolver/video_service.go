package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) VideoQuery() generated.VideoQueryResolver { return &videoQueryResolver{r} }

type videoQueryResolver struct{ *Resolver }

func (r *videoQueryResolver) Videos(ctx context.Context, obj *model.VideoQuery, page *int, pageSize *int, sort *string, qArg *string) (*model.VideoVideoList, error) {
	return r.Clients.VideoService.Videos(ctx, page, pageSize, sort, qArg)
}

func (r *videoQueryResolver) Video(ctx context.Context, obj *model.VideoQuery, id string) (*model.VideoVideo, error) {
	return r.Clients.VideoService.Video(ctx, id)
}

func (r *videoQueryResolver) Related(ctx context.Context, obj *model.VideoQuery, id string, limit *int) ([]*model.VideoVideo, error) {
	return r.Clients.VideoService.Related(ctx, id, limit)
}
