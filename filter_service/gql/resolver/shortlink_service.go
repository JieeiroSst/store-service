package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ShortlinkQuery() generated.ShortlinkQueryResolver {
	return &shortlinkQueryResolver{r}
}

type shortlinkQueryResolver struct{ *Resolver }

func (r *shortlinkQueryResolver) Links(ctx context.Context, obj *model.ShortlinkQuery, userID *string) ([]*model.ShortlinkLink, error) {
	return r.Clients.ShortlinkService.Links(ctx, userID)
}

func (r *shortlinkQueryResolver) Link(ctx context.Context, obj *model.ShortlinkQuery, id string, userID *string) (*model.ShortlinkLink, error) {
	return r.Clients.ShortlinkService.Link(ctx, id, userID)
}

func (r *shortlinkQueryResolver) AnalyticsOverview(ctx context.Context, obj *model.ShortlinkQuery, userID *string, days *int) (*model.ShortlinkAnalytics, error) {
	return r.Clients.ShortlinkService.AnalyticsOverview(ctx, userID, days)
}

func (r *shortlinkQueryResolver) LinkAnalytics(ctx context.Context, obj *model.ShortlinkQuery, linkID string, userID *string, days *int) (*model.ShortlinkAnalytics, error) {
	return r.Clients.ShortlinkService.LinkAnalytics(ctx, linkID, userID, days)
}

func (r *shortlinkQueryResolver) Resolve(ctx context.Context, obj *model.ShortlinkQuery, shortCode string) (*model.ShortlinkResolve, error) {
	return r.Clients.ShortlinkService.Resolve(ctx, shortCode)
}

func (r *shortlinkQueryResolver) ResolveWithTemplate(ctx context.Context, obj *model.ShortlinkQuery, templateSlug string, shortCode string) (*model.ShortlinkResolve, error) {
	return r.Clients.ShortlinkService.ResolveWithTemplate(ctx, templateSlug, shortCode)
}

func (r *shortlinkQueryResolver) Attribution(ctx context.Context, obj *model.ShortlinkQuery, fingerprint string) (*model.ShortlinkAttribution, error) {
	return r.Clients.ShortlinkService.Attribution(ctx, fingerprint)
}

func (r *shortlinkQueryResolver) SdkHealth(ctx context.Context, obj *model.ShortlinkQuery) (*model.ShortlinkSdkHealth, error) {
	return r.Clients.ShortlinkService.SdkHealth(ctx)
}

func (r *shortlinkQueryResolver) Webhooks(ctx context.Context, obj *model.ShortlinkQuery, userID *string) ([]*model.ShortlinkWebhook, error) {
	return r.Clients.ShortlinkService.Webhooks(ctx, userID)
}

func (r *shortlinkQueryResolver) Webhook(ctx context.Context, obj *model.ShortlinkQuery, id string, userID *string) (*model.ShortlinkWebhook, error) {
	return r.Clients.ShortlinkService.Webhook(ctx, id, userID)
}

func (r *shortlinkQueryResolver) Templates(ctx context.Context, obj *model.ShortlinkQuery, userID *string) ([]*model.ShortlinkTemplate, error) {
	return r.Clients.ShortlinkService.Templates(ctx, userID)
}

func (r *shortlinkQueryResolver) Template(ctx context.Context, obj *model.ShortlinkQuery, id string, userID *string) (*model.ShortlinkTemplate, error) {
	return r.Clients.ShortlinkService.Template(ctx, id, userID)
}

func (r *shortlinkQueryResolver) AppleAppSiteAssociation(ctx context.Context, obj *model.ShortlinkQuery) (map[string]interface{}, error) {
	return r.Clients.ShortlinkService.AppleAppSiteAssociation(ctx)
}

func (r *shortlinkQueryResolver) AssetLinks(ctx context.Context, obj *model.ShortlinkQuery) ([]map[string]interface{}, error) {
	return r.Clients.ShortlinkService.AssetLinks(ctx)
}
