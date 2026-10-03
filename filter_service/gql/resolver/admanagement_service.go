package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) AdmQuery() generated.AdmQueryResolver { return &admQueryResolver{r} }

type admQueryResolver struct{ *Resolver }

func (r *admQueryResolver) Campaigns(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAdCampaign, error) {
	return r.Clients.AdmanagementService.Campaigns(ctx)
}

func (r *admQueryResolver) Campaign(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAdCampaign, error) {
	return r.Clients.AdmanagementService.Campaign(ctx, id)
}

func (r *admQueryResolver) CampaignAds(ctx context.Context, obj *model.AdmQuery, id int) ([]*model.AdmAd, error) {
	return r.Clients.AdmanagementService.CampaignAds(ctx, id)
}

func (r *admQueryResolver) CampaignPerformance(ctx context.Context, obj *model.AdmQuery, id int, from *string, to *string) (*model.AdmCampaignPerformance, error) {
	return r.Clients.AdmanagementService.CampaignPerformance(ctx, id, from, to)
}

func (r *admQueryResolver) Categories(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAdCategory, error) {
	return r.Clients.AdmanagementService.Categories(ctx)
}

func (r *admQueryResolver) Category(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAdCategory, error) {
	return r.Clients.AdmanagementService.Category(ctx, id)
}

func (r *admQueryResolver) Positions(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAdPosition, error) {
	return r.Clients.AdmanagementService.Positions(ctx)
}

func (r *admQueryResolver) Position(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAdPosition, error) {
	return r.Clients.AdmanagementService.Position(ctx, id)
}

func (r *admQueryResolver) ServeAd(ctx context.Context, obj *model.AdmQuery, id int, sessionID *string, country *string, device *string, gender *string, referrerURL *string, pageURL *string, age *int, userID *int) (*model.AdmServedAd, error) {
	return r.Clients.AdmanagementService.ServeAd(ctx, id, sessionID, country, device, gender, referrerURL, pageURL, age, userID)
}

func (r *admQueryResolver) Ads(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAd, error) {
	return r.Clients.AdmanagementService.Ads(ctx)
}

func (r *admQueryResolver) Ad(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAd, error) {
	return r.Clients.AdmanagementService.Ad(ctx, id)
}

func (r *admQueryResolver) AdImpressions(ctx context.Context, obj *model.AdmQuery, id int) ([]*model.AdmAdImpression, error) {
	return r.Clients.AdmanagementService.AdImpressions(ctx, id)
}

func (r *admQueryResolver) AdClicks(ctx context.Context, obj *model.AdmQuery, id int) ([]*model.AdmAdClick, error) {
	return r.Clients.AdmanagementService.AdClicks(ctx, id)
}

func (r *admQueryResolver) PositionMappings(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAdPositionMapping, error) {
	return r.Clients.AdmanagementService.PositionMappings(ctx)
}

func (r *admQueryResolver) PositionMapping(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAdPositionMapping, error) {
	return r.Clients.AdmanagementService.PositionMapping(ctx, id)
}

func (r *admQueryResolver) Impressions(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAdImpression, error) {
	return r.Clients.AdmanagementService.Impressions(ctx)
}

func (r *admQueryResolver) Impression(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAdImpression, error) {
	return r.Clients.AdmanagementService.Impression(ctx, id)
}

func (r *admQueryResolver) Clicks(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAdClick, error) {
	return r.Clients.AdmanagementService.Clicks(ctx)
}

func (r *admQueryResolver) Click(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAdClick, error) {
	return r.Clients.AdmanagementService.Click(ctx, id)
}

func (r *admQueryResolver) TargetingRules(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAdTargetingRule, error) {
	return r.Clients.AdmanagementService.TargetingRules(ctx)
}

func (r *admQueryResolver) TargetingRule(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAdTargetingRule, error) {
	return r.Clients.AdmanagementService.TargetingRule(ctx, id)
}

func (r *admQueryResolver) PerformanceSummaries(ctx context.Context, obj *model.AdmQuery) ([]*model.AdmAdPerformanceSummary, error) {
	return r.Clients.AdmanagementService.PerformanceSummaries(ctx)
}

func (r *admQueryResolver) PerformanceSummary(ctx context.Context, obj *model.AdmQuery, id int) (*model.AdmAdPerformanceSummary, error) {
	return r.Clients.AdmanagementService.PerformanceSummary(ctx, id)
}
