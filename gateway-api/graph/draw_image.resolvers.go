package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) DrawImageCreateCollage(ctx context.Context, files []*graphql.Upload, center *graphql.Upload, layout *string, columns *int) (*model.DrawImageCollage, error) {
	return restMultipart[*model.DrawImageCollage](ctx, r.Proxy, "draw-image-service", rest.Request{Method: "POST", Path: "/upload", Query: restQuery("layout", layout, "columns", columns)}, map[string]any{}, uploadFiles("files", files, "center", center), "")
}

func (r *queryResolver) DrawImageCollage(ctx context.Context, id string) (*model.DrawImageCollage, error) {
	return restCall[*model.DrawImageCollage](ctx, r.Proxy, "draw-image-service", rest.Request{Method: "GET", Path: rest.P("/images/%v/info", id)}, "")
}
