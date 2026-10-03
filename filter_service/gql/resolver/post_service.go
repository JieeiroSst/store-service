package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) PostQuery() generated.PostQueryResolver { return &postQueryResolver{r} }

type postQueryResolver struct{ *Resolver }

func (r *postQueryResolver) Posts(ctx context.Context, obj *model.PostQuery, cursor *string, limit *int) (*model.PostPostPage, error) {
	return r.Clients.PostService.Posts(ctx, cursor, limit)
}

func (r *postQueryResolver) Post(ctx context.Context, obj *model.PostQuery, id string) (*model.PostPost, error) {
	return r.Clients.PostService.Post(ctx, id)
}

func (r *postQueryResolver) Categories(ctx context.Context, obj *model.PostQuery, cursor *string, limit *int) (*model.PostCategoryPage, error) {
	return r.Clients.PostService.Categories(ctx, cursor, limit)
}

func (r *postQueryResolver) Category(ctx context.Context, obj *model.PostQuery, id string) (*model.PostCategory, error) {
	return r.Clients.PostService.Category(ctx, id)
}
