package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) VideoUpload(ctx context.Context, file graphql.Upload, title *string, description *string) (*model.VideoVideo, error) {
	return restMultipart[*model.VideoVideo](ctx, r.Proxy, "video-service", rest.Request{Method: "POST", Path: "/api/videos"}, map[string]any{"title": title, "description": description}, uploadFiles("file", file), "")
}

func (r *mutationResolver) VideoView(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "video-service", rest.Request{Method: "POST", Path: rest.P("/api/videos/%v/view", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) VideoDelete(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "video-service", rest.Request{Method: "DELETE", Path: rest.P("/api/videos/%v", id)}, "")
}

func (r *queryResolver) Videos(ctx context.Context, page *int, pageSize *int, sort *string, q *string) (*model.VideoVideoList, error) {
	return restCall[*model.VideoVideoList](ctx, r.Proxy, "video-service", rest.Request{Method: "GET", Path: "/api/videos", Query: restQuery("page", page, "page_size", pageSize, "sort", sort, "q", q)}, "")
}

func (r *queryResolver) Video(ctx context.Context, id string) (*model.VideoVideo, error) {
	return restCall[*model.VideoVideo](ctx, r.Proxy, "video-service", rest.Request{Method: "GET", Path: rest.P("/api/videos/%v", id)}, "")
}

func (r *queryResolver) VideoRelated(ctx context.Context, id string, limit *int) ([]*model.VideoVideo, error) {
	return restCall[[]*model.VideoVideo](ctx, r.Proxy, "video-service", rest.Request{Method: "GET", Path: rest.P("/api/videos/%v/related", id), Query: restQuery("limit", limit)}, "items")
}
