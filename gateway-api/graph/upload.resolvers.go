package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) UploadFile(ctx context.Context, file graphql.Upload, receiverID *string) (*model.UploadFile, error) {
	return restMultipart[*model.UploadFile](ctx, r.Proxy, "upload-service", rest.Request{Method: "POST", Path: "/api/v1/upload", Query: restQuery("receiver_id", receiverID)}, map[string]any{}, uploadFiles("file", file), "")
}

func (r *mutationResolver) UploadReplaceFile(ctx context.Context, id string, file graphql.Upload) (*model.UploadFile, error) {
	return restMultipart[*model.UploadFile](ctx, r.Proxy, "upload-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/upload/%v", id)}, map[string]any{}, uploadFiles("file", file), "")
}

func (r *mutationResolver) UploadDeleteFile(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "upload-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/upload/%v", id)}, "")
}

func (r *queryResolver) UploadFiles(ctx context.Context, receiverID *string, limit *int, offset *int) (*model.UploadFileList, error) {
	return restCall[*model.UploadFileList](ctx, r.Proxy, "upload-service", rest.Request{Method: "GET", Path: "/api/v1/upload", Query: restQuery("receiver_id", receiverID, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) UploadFile(ctx context.Context, id string) (*model.UploadFile, error) {
	return restCall[*model.UploadFile](ctx, r.Proxy, "upload-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/upload/%v", id)}, "")
}
