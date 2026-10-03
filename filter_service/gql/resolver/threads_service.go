package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ThreadsQuery() generated.ThreadsQueryResolver { return &threadsQueryResolver{r} }

type threadsQueryResolver struct{ *Resolver }

func (r *threadsQueryResolver) Posts(ctx context.Context, obj *model.ThreadsQuery, authorID *string, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	return r.Clients.ThreadsService.Posts(ctx, authorID, cursor, limit)
}

func (r *threadsQueryResolver) Post(ctx context.Context, obj *model.ThreadsQuery, id string) (*model.ThreadsPost, error) {
	return r.Clients.ThreadsService.Post(ctx, id)
}

func (r *threadsQueryResolver) Comments(ctx context.Context, obj *model.ThreadsQuery, id string, cursor *string, limit *int) (*model.ThreadsCommentPage, error) {
	return r.Clients.ThreadsService.Comments(ctx, id, cursor, limit)
}

func (r *threadsQueryResolver) Followers(ctx context.Context, obj *model.ThreadsQuery, id string, cursor *string, limit *int) (*model.ThreadsAuthorPage, error) {
	return r.Clients.ThreadsService.Followers(ctx, id, cursor, limit)
}

func (r *threadsQueryResolver) Following(ctx context.Context, obj *model.ThreadsQuery, id string, cursor *string, limit *int) (*model.ThreadsAuthorPage, error) {
	return r.Clients.ThreadsService.Following(ctx, id, cursor, limit)
}

func (r *threadsQueryResolver) HomeFeed(ctx context.Context, obj *model.ThreadsQuery, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	return r.Clients.ThreadsService.HomeFeed(ctx, cursor, limit)
}

func (r *threadsQueryResolver) Bookmarks(ctx context.Context, obj *model.ThreadsQuery, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	return r.Clients.ThreadsService.Bookmarks(ctx, cursor, limit)
}

func (r *threadsQueryResolver) TrendingTags(ctx context.Context, obj *model.ThreadsQuery, limit *int) ([]*model.ThreadsTagCount, error) {
	return r.Clients.ThreadsService.TrendingTags(ctx, limit)
}
