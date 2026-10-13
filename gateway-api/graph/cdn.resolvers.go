package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) CdnPresign(ctx context.Context, input model.CdnPresignInput) (*model.CdnPresign, error) {
	return restCall[*model.CdnPresign](ctx, r.Proxy, "cdn-service", rest.Request{Method: "POST", Path: "/api/v1/files/presign", Body: input}, "")
}

func (r *mutationResolver) CdnConfirm(ctx context.Context, id string) (*model.CdnFile, error) {
	return restCall[*model.CdnFile](ctx, r.Proxy, "cdn-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/files/%v/confirm", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CdnDeleteFile(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "cdn-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/files/%v", id)}, "")
}

func (r *queryResolver) CdnFiles(ctx context.Context, limit *int, offset *int) (*model.CdnFileList, error) {
	return restCall[*model.CdnFileList](ctx, r.Proxy, "cdn-service", rest.Request{Method: "GET", Path: "/api/v1/files", Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) CdnFile(ctx context.Context, id string) (*model.CdnFile, error) {
	return restCall[*model.CdnFile](ctx, r.Proxy, "cdn-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/files/%v", id)}, "")
}

func (r *queryResolver) CdnDownloadURL(ctx context.Context, id string, direct *bool) (*model.CdnDownload, error) {
	return restCall[*model.CdnDownload](ctx, r.Proxy, "cdn-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/files/%v/download", id), Query: restQuery("direct", direct)}, "")
}
