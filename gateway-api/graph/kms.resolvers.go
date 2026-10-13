package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) KmsCreateKey(ctx context.Context, input model.KmsCreateKeyInput) (*model.KmsKey, error) {
	return restCall[*model.KmsKey](ctx, r.Proxy, "kms-service", rest.Request{Method: "POST", Path: "/api/v1/keys", Body: input}, "")
}

func (r *mutationResolver) KmsRotateKey(ctx context.Context, id string, force *bool) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "kms-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/keys/%v/rotate", id), Body: map[string]any{"force": force}}, "message")
}

func (r *mutationResolver) KmsDeleteKey(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "kms-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/keys/%v", id)}, "")
}

func (r *queryResolver) KmsKeys(ctx context.Context) ([]*model.KmsKey, error) {
	return restCall[[]*model.KmsKey](ctx, r.Proxy, "kms-service", rest.Request{Method: "GET", Path: "/api/v1/keys"}, "")
}

func (r *queryResolver) KmsKey(ctx context.Context, id string) (*model.KmsKey, error) {
	return restCall[*model.KmsKey](ctx, r.Proxy, "kms-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/keys/%v", id)}, "")
}

func (r *queryResolver) KmsKeyForUse(ctx context.Context, id string) (*model.KmsKeyUse, error) {
	return restCall[*model.KmsKeyUse](ctx, r.Proxy, "kms-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/keys/%v/use", id)}, "")
}

func (r *queryResolver) KmsKeyUsageStats(ctx context.Context, id string) (*model.KmsKeyStats, error) {
	return restCall[*model.KmsKeyStats](ctx, r.Proxy, "kms-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/keys/%v/stats", id)}, "")
}

func (r *queryResolver) KmsKeyAuditLogs(ctx context.Context, id string, limit *int, offset *int) (*model.KmsKeyAuditLogs, error) {
	return restCall[*model.KmsKeyAuditLogs](ctx, r.Proxy, "kms-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/keys/%v/audit", id), Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) KmsAuditLogs(ctx context.Context, limit *int, offset *int) (*model.KmsAuditLogs, error) {
	return restCall[*model.KmsAuditLogs](ctx, r.Proxy, "kms-service", rest.Request{Method: "GET", Path: "/api/v1/audit/logs", Query: restQuery("limit", limit, "offset", offset)}, "")
}
