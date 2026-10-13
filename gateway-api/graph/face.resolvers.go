package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) FaceReportPositions(ctx context.Context, input model.FacePositionsInput) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "face-recognition-service", rest.Request{Method: "POST", Path: "/face/position", Body: input}, "ok")
}

func (r *mutationResolver) FaceUploadFile(ctx context.Context, avatar graphql.Upload) (*string, error) {
	return restMultipart[*string](ctx, r.Proxy, "face-recognition-service", rest.Request{Method: "POST", Path: "/upload-single-file"}, map[string]any{}, uploadFiles("avatar", avatar), "")
}

func (r *mutationResolver) FaceUploadFiles(ctx context.Context, avatar *graphql.Upload, boop []*graphql.Upload) (*string, error) {
	return restMultipart[*string](ctx, r.Proxy, "face-recognition-service", rest.Request{Method: "POST", Path: "/upload-multiple-files"}, map[string]any{}, uploadFiles("avatar", avatar, "boop", boop), "")
}

func (r *queryResolver) FacePositions(ctx context.Context) (*model.FacePositions, error) {
	return restCall[*model.FacePositions](ctx, r.Proxy, "face-recognition-service", rest.Request{Method: "GET", Path: "/face/position"}, "")
}
