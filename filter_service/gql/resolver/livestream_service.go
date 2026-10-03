package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) LiveQuery() generated.LiveQueryResolver { return &liveQueryResolver{r} }

type liveQueryResolver struct{ *Resolver }

func (r *liveQueryResolver) Rooms(ctx context.Context, obj *model.LiveQuery, live *bool) ([]*model.LiveRoom, error) {
	return r.Clients.LivestreamService.Rooms(ctx, live)
}

func (r *liveQueryResolver) Room(ctx context.Context, obj *model.LiveQuery, id string) (*model.LiveRoom, error) {
	return r.Clients.LivestreamService.Room(ctx, id)
}

func (r *liveQueryResolver) ActiveStream(ctx context.Context, obj *model.LiveQuery, id string) (*model.LiveStream, error) {
	return r.Clients.LivestreamService.ActiveStream(ctx, id)
}

func (r *liveQueryResolver) Recordings(ctx context.Context, obj *model.LiveQuery, id string) ([]*model.LiveRecording, error) {
	return r.Clients.LivestreamService.Recordings(ctx, id)
}

func (r *liveQueryResolver) ViewerCount(ctx context.Context, obj *model.LiveQuery, id string) (*model.LiveViewerCount, error) {
	return r.Clients.LivestreamService.ViewerCount(ctx, id)
}

func (r *liveQueryResolver) Playback(ctx context.Context, obj *model.LiveQuery, id string) (*model.LivePlaybackInfo, error) {
	return r.Clients.LivestreamService.Playback(ctx, id)
}
