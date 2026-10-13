package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) CreateMyQRCode(ctx context.Context, input model.QRGenerateInput) (*model.QRQRCode, error) {
	return restCall[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "POST", Path: "/api/v1/qr", Body: withField(input, "created_by", userIDOf(ctx))}, "data")
}

func (r *mutationResolver) UpdateMyQRCode(ctx context.Context, id string, input model.QRUpdateInput) (*model.QRQRCode, error) {
	return restOwned[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/qr/%v", id), Body: input}, "data", "created_by", &rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v", id)}, "data")
}

func (r *mutationResolver) UpdateMyQRContent(ctx context.Context, id string, content *string, redirectURL *string) (*model.QRQRCode, error) {
	return restOwned[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/qr/%v/content", id), Body: map[string]any{"content": content, "redirect_url": redirectURL}}, "data", "created_by", &rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v", id)}, "data")
}

func (r *mutationResolver) DeleteMyQRCode(ctx context.Context, id string) (bool, error) {
	return restOwned[bool](ctx, r.Proxy, "qr-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/qr/%v", id)}, "", "created_by", &rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v", id)}, "data")
}

func (r *mutationResolver) RegenerateMyQRCode(ctx context.Context, id string) (*model.QRQRCode, error) {
	return restOwned[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/qr/%v/regenerate", id), Body: map[string]any{}}, "data", "created_by", &rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v", id)}, "data")
}

func (r *mutationResolver) QRGenerate(ctx context.Context, input model.QRGenerateInput) (*model.QRQRCode, error) {
	return restCall[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "POST", Path: "/api/v1/qr", Body: withField(input, "created_by", userIDOf(ctx))}, "data")
}

func (r *mutationResolver) QRUpdate(ctx context.Context, id string, input model.QRUpdateInput) (*model.QRQRCode, error) {
	return restCall[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/qr/%v", id), Body: input}, "data")
}

func (r *mutationResolver) QRUpdateContent(ctx context.Context, id string, content *string, redirectURL *string) (*model.QRQRCode, error) {
	return restCall[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/qr/%v/content", id), Body: map[string]any{"content": content, "redirect_url": redirectURL}}, "data")
}

func (r *mutationResolver) QRDelete(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "qr-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/qr/%v", id)}, "")
}

func (r *mutationResolver) QRRegenerate(ctx context.Context, id string) (*model.QRQRCode, error) {
	return restCall[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/qr/%v/regenerate", id), Body: map[string]any{}}, "data")
}

func (r *queryResolver) QRCodes(ctx context.Context, page *int, limit *int, status *string, typeArg *string, search *string, createdBy *string) (*model.QRQRListResponse, error) {
	return restCall[*model.QRQRListResponse](ctx, r.Proxy, "qr-service", rest.Request{Method: "GET", Path: "/api/v1/qr", Query: restQuery("page", page, "limit", limit, "status", status, "type", typeArg, "search", search, "created_by", createdBy)}, "data")
}

func (r *queryResolver) QRCode(ctx context.Context, id string) (*model.QRQRCode, error) {
	return restCall[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v", id)}, "data")
}

func (r *queryResolver) QRScanHistory(ctx context.Context, id string, page *int, limit *int) (*model.QRScanHistoryResponse, error) {
	return restCall[*model.QRScanHistoryResponse](ctx, r.Proxy, "qr-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v/history", id), Query: restQuery("page", page, "limit", limit)}, "data")
}

func (r *queryResolver) QRScanStats(ctx context.Context, id string) (*model.QRScanStats, error) {
	return restCall[*model.QRScanStats](ctx, r.Proxy, "qr-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v/stats", id)}, "data")
}

func (r *queryResolver) MyQRCodes(ctx context.Context, page *int, limit *int, status *string, typeArg *string, search *string) (*model.QRQRListResponse, error) {
	return restCall[*model.QRQRListResponse](ctx, r.Proxy, "qr-service", rest.Request{Method: "GET", Path: "/api/v1/qr", Query: restQuery("page", page, "limit", limit, "status", status, "type", typeArg, "search", search, "created_by", userIDOf(ctx))}, "data")
}

func (r *queryResolver) MyQRCode(ctx context.Context, id string) (*model.QRQRCode, error) {
	return restOwned[*model.QRQRCode](ctx, r.Proxy, "qr-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v", id)}, "data", "created_by", nil, "")
}

func (r *queryResolver) MyQRScanStats(ctx context.Context, id string) (*model.QRScanStats, error) {
	return restOwned[*model.QRScanStats](ctx, r.Proxy, "qr-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v/stats", id)}, "data", "created_by", &rest.Request{Method: "GET", Path: rest.P("/api/v1/qr/%v", id)}, "data")
}
