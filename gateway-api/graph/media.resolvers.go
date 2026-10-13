package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) MediaUploadVideo(ctx context.Context, video graphql.Upload) (*string, error) {
	return restMultipart[*string](ctx, r.Proxy, "media-service", rest.Request{Method: "POST", Path: "/api/internal/v1/videos/upload"}, map[string]any{}, uploadFiles("video", video), "data")
}

func (r *mutationResolver) MediaSaveVideo(ctx context.Context, video map[string]any, tag map[string]any) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "media-service", rest.Request{Method: "POST", Path: "/api/internal/v1/videos/", Body: map[string]any{"video": video, "tag": tag}}, "")
}

func (r *queryResolver) MediaVideo(ctx context.Context, videoID int, userID int) (*model.MediaVideo, error) {
	return restCall[*model.MediaVideo](ctx, r.Proxy, "media-service", rest.Request{Method: "GET", Path: rest.P("/api/internal/v1/videos/%v/user/%v", videoID, userID)}, "data")
}

func (r *queryResolver) MediaSearchVideos(ctx context.Context, query *string, page *int, size *int) (*model.MediaSearchResult, error) {
	return restCall[*model.MediaSearchResult](ctx, r.Proxy, "media-service", rest.Request{Method: "POST", Path: "/api/internal/v1/videos/search", Body: map[string]any{"query": query, "page": page, "size": size}}, "data")
}
