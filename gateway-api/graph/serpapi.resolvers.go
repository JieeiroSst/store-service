package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) SerpUploadImage(ctx context.Context, image graphql.Upload) (any, error) {
	return restMultipart[any](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "POST", Path: "/api/v1/images"}, map[string]any{}, uploadFiles("image", image), "")
}

func (r *queryResolver) SerpEngines(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "GET", Path: "/api/v1/engines"}, "")
}

func (r *queryResolver) SerpEngine(ctx context.Context, engine string) (any, error) {
	return restCall[any](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/engines/%v", engine)}, "")
}

func (r *queryResolver) SerpSearch(ctx context.Context, engine *string, params map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "GET", Path: "/api/v1/search", Query: restQuery("engine", engine, "params", params)}, "")
}

func (r *queryResolver) SerpSearchBody(ctx context.Context, input model.SerpSearchInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "POST", Path: "/api/v1/search", Body: input}, "")
}

func (r *queryResolver) SerpSearchBatch(ctx context.Context, searches []*model.SerpSearchInput) (*model.SerpBatch, error) {
	return restCall[*model.SerpBatch](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "POST", Path: "/api/v1/search/batch", Body: map[string]any{"searches": searches}}, "")
}

func (r *queryResolver) SerpArchive(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/searches/%v", id)}, "")
}

func (r *queryResolver) SerpAccount(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "GET", Path: "/api/v1/account"}, "")
}

func (r *queryResolver) SerpLocations(ctx context.Context, q *string, limit *int) (any, error) {
	return restCall[any](ctx, r.Proxy, "serpapi-service", rest.Request{Method: "GET", Path: "/api/v1/locations", Query: restQuery("q", q, "limit", limit)}, "")
}
