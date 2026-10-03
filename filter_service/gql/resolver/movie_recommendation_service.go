package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) MovieQuery() generated.MovieQueryResolver { return &movieQueryResolver{r} }

type movieQueryResolver struct{ *Resolver }

func (r *movieQueryResolver) ForUser(ctx context.Context, obj *model.MovieQuery, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return r.Clients.MovieRecommendationService.ForUser(ctx, userID, page, pageSize, snapshot)
}

func (r *movieQueryResolver) Home(ctx context.Context, obj *model.MovieQuery, userID string) (*model.MovieHome, error) {
	return r.Clients.MovieRecommendationService.Home(ctx, userID)
}

func (r *movieQueryResolver) ContinueWatching(ctx context.Context, obj *model.MovieQuery, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return r.Clients.MovieRecommendationService.ContinueWatching(ctx, userID, page, pageSize, snapshot)
}

func (r *movieQueryResolver) History(ctx context.Context, obj *model.MovieQuery, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return r.Clients.MovieRecommendationService.History(ctx, userID, page, pageSize, snapshot)
}

func (r *movieQueryResolver) NewReleases(ctx context.Context, obj *model.MovieQuery, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return r.Clients.MovieRecommendationService.NewReleases(ctx, page, pageSize, snapshot)
}

func (r *movieQueryResolver) Similar(ctx context.Context, obj *model.MovieQuery, id string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return r.Clients.MovieRecommendationService.Similar(ctx, id, page, pageSize, snapshot)
}

func (r *movieQueryResolver) Trending(ctx context.Context, obj *model.MovieQuery, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return r.Clients.MovieRecommendationService.Trending(ctx, page, pageSize, snapshot)
}
