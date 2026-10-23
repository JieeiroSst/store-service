package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) MovieRecordEvent(ctx context.Context, videoID string, typeArg string, value *float64) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "POST", Path: "/api/recommendations/events", Body: withField(map[string]any{"video_id": videoID, "type": typeArg, "value": value}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) MovieRemoveFromMyHistory(ctx context.Context, videoID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "DELETE", Path: rest.P("/api/recommendations/users/%v/history/%v", userIDOf(ctx), videoID)}, "")
}

func (r *mutationResolver) MovieUpsertVideo(ctx context.Context, id string, input model.MovieVideoInput) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "PUT", Path: rest.P("/api/recommendations/videos/%v", id), Body: input}, "")
}

func (r *mutationResolver) MovieDeleteVideo(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "DELETE", Path: rest.P("/api/recommendations/videos/%v", id)}, "")
}

func (r *queryResolver) MovieForUser(ctx context.Context, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/users/%v", userID), Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}

func (r *queryResolver) MovieHome(ctx context.Context, userID string) (*model.MovieHome, error) {
	return restCall[*model.MovieHome](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/users/%v/home", userID)}, "")
}

func (r *queryResolver) MovieContinueWatching(ctx context.Context, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/users/%v/continue-watching", userID), Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}

func (r *queryResolver) MovieHistory(ctx context.Context, userID string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/users/%v/history", userID), Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}

func (r *queryResolver) MovieNewReleases(ctx context.Context, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: "/api/recommendations/new", Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}

func (r *queryResolver) MovieVideo(ctx context.Context, id string) (*model.MovieVideo, error) {
	return restCall[*model.MovieVideo](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/videos/%v", id)}, "")
}

func (r *queryResolver) MovieSimilar(ctx context.Context, id string, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/videos/%v/similar", id), Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}

func (r *queryResolver) MovieTrending(ctx context.Context, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: "/api/recommendations/trending", Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}

func (r *queryResolver) MyMovieForUser(ctx context.Context, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/users/%v", userIDOf(ctx)), Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}

func (r *queryResolver) MyMovieHome(ctx context.Context) (*model.MovieHome, error) {
	return restCall[*model.MovieHome](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/users/%v/home", userIDOf(ctx))}, "")
}

func (r *queryResolver) MyMovieContinueWatching(ctx context.Context, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/users/%v/continue-watching", userIDOf(ctx)), Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}

func (r *queryResolver) MyMovieHistory(ctx context.Context, page *int, pageSize *int, snapshot *string) (*model.MoviePage, error) {
	return restCall[*model.MoviePage](ctx, r.Proxy, "movie-recommendation-service", rest.Request{Method: "GET", Path: rest.P("/api/recommendations/users/%v/history", userIDOf(ctx)), Query: restQuery("page", page, "page_size", pageSize, "snapshot", snapshot)}, "")
}
