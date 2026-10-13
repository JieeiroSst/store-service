package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) PhotoCreateComposition(ctx context.Context, images []*graphql.Upload, urls []string, layout *string) (*model.PhotoComposition, error) {
	return restMultipart[*model.PhotoComposition](ctx, r.Proxy, "photo-service", rest.Request{Method: "POST", Path: "/v1/compositions"}, map[string]any{"urls": urls, "layout": layout}, uploadFiles("images", images), "")
}

func (r *queryResolver) PhotoComposition(ctx context.Context, id string) (*model.PhotoComposition, error) {
	return restCall[*model.PhotoComposition](ctx, r.Proxy, "photo-service", rest.Request{Method: "GET", Path: rest.P("/v1/compositions/%v", id)}, "")
}
