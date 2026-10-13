package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ShortlinkCreateLink(ctx context.Context, input model.ShortlinkCreateLinkInput) (*model.ShortlinkLink, error) {
	return restCall[*model.ShortlinkLink](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "POST", Path: "/api/links", Body: input}, "")
}

func (r *mutationResolver) ShortlinkUpdateLink(ctx context.Context, id string, input model.ShortlinkUpdateLinkInput) (*model.ShortlinkLink, error) {
	return restCall[*model.ShortlinkLink](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "PUT", Path: rest.P("/api/links/%v", id), Body: input}, "")
}

func (r *mutationResolver) ShortlinkDeleteLink(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "DELETE", Path: rest.P("/api/links/%v", id)}, "")
}

func (r *mutationResolver) ShortlinkDuplicateLink(ctx context.Context, id string) (*model.ShortlinkLink, error) {
	return restCall[*model.ShortlinkLink](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "POST", Path: rest.P("/api/links/%v/duplicate", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ShortlinkSdkInstall(ctx context.Context, input model.ShortlinkInstallInput) (*model.ShortlinkAttribution, error) {
	return restCall[*model.ShortlinkAttribution](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "POST", Path: "/api/sdk/v1/install", Body: input}, "")
}

func (r *mutationResolver) ShortlinkSdkEvent(ctx context.Context, input model.ShortlinkEventInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "POST", Path: "/api/sdk/v1/event", Body: input}, "")
}

func (r *mutationResolver) ShortlinkCreateWebhook(ctx context.Context, input model.ShortlinkWebhookInput) (*model.ShortlinkWebhook, error) {
	return restCall[*model.ShortlinkWebhook](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "POST", Path: "/api/webhooks", Body: input}, "")
}

func (r *mutationResolver) ShortlinkUpdateWebhook(ctx context.Context, id string, input model.ShortlinkWebhookInput) (*model.ShortlinkWebhook, error) {
	return restCall[*model.ShortlinkWebhook](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "PUT", Path: rest.P("/api/webhooks/%v", id), Body: input}, "")
}

func (r *mutationResolver) ShortlinkDeleteWebhook(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "DELETE", Path: rest.P("/api/webhooks/%v", id)}, "")
}

func (r *mutationResolver) ShortlinkTestWebhook(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "POST", Path: rest.P("/api/webhooks/%v/test", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ShortlinkCreateTemplate(ctx context.Context, input model.ShortlinkTemplateInput) (*model.ShortlinkTemplate, error) {
	return restCall[*model.ShortlinkTemplate](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "POST", Path: "/api/templates", Body: input}, "")
}

func (r *mutationResolver) ShortlinkUpdateTemplate(ctx context.Context, id string, input model.ShortlinkTemplateInput) (*model.ShortlinkTemplate, error) {
	return restCall[*model.ShortlinkTemplate](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "PUT", Path: rest.P("/api/templates/%v", id), Body: input}, "")
}

func (r *mutationResolver) ShortlinkDeleteTemplate(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "DELETE", Path: rest.P("/api/templates/%v", id)}, "")
}

func (r *mutationResolver) ShortlinkSetDefaultTemplate(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "PUT", Path: rest.P("/api/templates/%v/set-default", id), Body: map[string]any{}}, "")
}

func (r *queryResolver) ShortlinkLinks(ctx context.Context, userID *string) ([]*model.ShortlinkLink, error) {
	return restCall[[]*model.ShortlinkLink](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: "/api/links", Query: restQuery("userId", userID)}, "")
}

func (r *queryResolver) ShortlinkLink(ctx context.Context, id string, userID *string) (*model.ShortlinkLink, error) {
	return restCall[*model.ShortlinkLink](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: rest.P("/api/links/%v", id), Query: restQuery("userId", userID)}, "")
}

func (r *queryResolver) ShortlinkAnalyticsOverview(ctx context.Context, userID *string, days *int) (*model.ShortlinkAnalytics, error) {
	return restCall[*model.ShortlinkAnalytics](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: "/api/analytics/overview", Query: restQuery("userId", userID, "days", days)}, "")
}

func (r *queryResolver) ShortlinkLinkAnalytics(ctx context.Context, linkID string, userID *string, days *int) (*model.ShortlinkAnalytics, error) {
	return restCall[*model.ShortlinkAnalytics](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: rest.P("/api/analytics/links/%v", linkID), Query: restQuery("userId", userID, "days", days)}, "")
}

func (r *queryResolver) ShortlinkResolve(ctx context.Context, shortCode string) (*model.ShortlinkResolve, error) {
	return restCall[*model.ShortlinkResolve](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: rest.P("/api/sdk/v1/resolve/%v", shortCode)}, "")
}

func (r *queryResolver) ShortlinkResolveWithTemplate(ctx context.Context, templateSlug string, shortCode string) (*model.ShortlinkResolve, error) {
	return restCall[*model.ShortlinkResolve](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: rest.P("/api/sdk/v1/resolve/%v/%v", templateSlug, shortCode)}, "")
}

func (r *queryResolver) ShortlinkAttribution(ctx context.Context, fingerprint string) (*model.ShortlinkAttribution, error) {
	return restCall[*model.ShortlinkAttribution](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: rest.P("/api/sdk/v1/attribution/%v", fingerprint)}, "")
}

func (r *queryResolver) ShortlinkSdkHealth(ctx context.Context) (*model.ShortlinkSdkHealth, error) {
	return restCall[*model.ShortlinkSdkHealth](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: "/api/sdk/v1/health"}, "")
}

func (r *queryResolver) ShortlinkWebhooks(ctx context.Context, userID *string) ([]*model.ShortlinkWebhook, error) {
	return restCall[[]*model.ShortlinkWebhook](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: "/api/webhooks", Query: restQuery("userId", userID)}, "")
}

func (r *queryResolver) ShortlinkWebhook(ctx context.Context, id string, userID *string) (*model.ShortlinkWebhook, error) {
	return restCall[*model.ShortlinkWebhook](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: rest.P("/api/webhooks/%v", id), Query: restQuery("userId", userID)}, "")
}

func (r *queryResolver) ShortlinkTemplates(ctx context.Context, userID *string) ([]*model.ShortlinkTemplate, error) {
	return restCall[[]*model.ShortlinkTemplate](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: "/api/templates", Query: restQuery("userId", userID)}, "")
}

func (r *queryResolver) ShortlinkTemplate(ctx context.Context, id string, userID *string) (*model.ShortlinkTemplate, error) {
	return restCall[*model.ShortlinkTemplate](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: rest.P("/api/templates/%v", id), Query: restQuery("userId", userID)}, "")
}

func (r *queryResolver) ShortlinkAppleAppSiteAssociation(ctx context.Context) (map[string]any, error) {
	return restCall[map[string]any](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: "/.well-known/apple-app-site-association"}, "")
}

func (r *queryResolver) ShortlinkAssetLinks(ctx context.Context) ([]map[string]any, error) {
	return restCall[[]map[string]any](ctx, r.Proxy, "shortlink-service", rest.Request{Method: "GET", Path: "/.well-known/assetlinks.json"}, "")
}
