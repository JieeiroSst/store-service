package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) AdCreateCampaign(ctx context.Context, input model.AdCampaignInput) (*model.AdCampaign, error) {
	return restCall[*model.AdCampaign](ctx, r.Proxy, "admanagement", rest.Request{Method: "POST", Path: "/campaigns", Body: input}, "")
}

func (r *mutationResolver) AdUpdateCampaign(ctx context.Context, id int, input model.AdCampaignInput) (*model.AdCampaign, error) {
	return restCall[*model.AdCampaign](ctx, r.Proxy, "admanagement", rest.Request{Method: "PUT", Path: rest.P("/campaigns/%v", id), Body: input}, "")
}

func (r *mutationResolver) AdDeleteCampaign(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "admanagement", rest.Request{Method: "DELETE", Path: rest.P("/campaigns/%v", id)}, "")
}

func (r *mutationResolver) AdChangeCampaignStatus(ctx context.Context, id int, input model.AdCampaignStatusInput) (*model.AdCampaign, error) {
	return restCall[*model.AdCampaign](ctx, r.Proxy, "admanagement", rest.Request{Method: "PATCH", Path: rest.P("/campaigns/%v", id), Body: input}, "")
}

func (r *mutationResolver) AdCreateCategory(ctx context.Context, input model.AdCategoryInput) (*model.AdCategory, error) {
	return restCall[*model.AdCategory](ctx, r.Proxy, "admanagement", rest.Request{Method: "POST", Path: "/categories", Body: input}, "")
}

func (r *mutationResolver) AdUpdateCategory(ctx context.Context, id int, input model.AdCategoryInput) (*model.AdCategory, error) {
	return restCall[*model.AdCategory](ctx, r.Proxy, "admanagement", rest.Request{Method: "PUT", Path: rest.P("/categories/%v", id), Body: input}, "")
}

func (r *mutationResolver) AdDeleteCategory(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "admanagement", rest.Request{Method: "DELETE", Path: rest.P("/categories/%v", id)}, "")
}

func (r *mutationResolver) AdCreatePosition(ctx context.Context, input model.AdPositionInput) (*model.AdPosition, error) {
	return restCall[*model.AdPosition](ctx, r.Proxy, "admanagement", rest.Request{Method: "POST", Path: "/positions", Body: input}, "")
}

func (r *mutationResolver) AdUpdatePosition(ctx context.Context, id int, input model.AdPositionInput) (*model.AdPosition, error) {
	return restCall[*model.AdPosition](ctx, r.Proxy, "admanagement", rest.Request{Method: "PUT", Path: rest.P("/positions/%v", id), Body: input}, "")
}

func (r *mutationResolver) AdDeletePosition(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "admanagement", rest.Request{Method: "DELETE", Path: rest.P("/positions/%v", id)}, "")
}

func (r *mutationResolver) AdCreateAd(ctx context.Context, input model.AdAdInput) (*model.AdAd, error) {
	return restCall[*model.AdAd](ctx, r.Proxy, "admanagement", rest.Request{Method: "POST", Path: "/ads", Body: input}, "")
}

func (r *mutationResolver) AdUpdateAd(ctx context.Context, id int, input model.AdAdInput) (*model.AdAd, error) {
	return restCall[*model.AdAd](ctx, r.Proxy, "admanagement", rest.Request{Method: "PUT", Path: rest.P("/ads/%v", id), Body: input}, "")
}

func (r *mutationResolver) AdDeleteAd(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "admanagement", rest.Request{Method: "DELETE", Path: rest.P("/ads/%v", id)}, "")
}

func (r *mutationResolver) AdRecomputePerformance(ctx context.Context, id int, date *string) (*model.AdPerformanceSummary, error) {
	return restCall[*model.AdPerformanceSummary](ctx, r.Proxy, "admanagement", rest.Request{Method: "POST", Path: rest.P("/ads/%v/performance/recompute", id), Body: map[string]any{"date": date}}, "")
}

func (r *mutationResolver) AdCreatePositionMapping(ctx context.Context, input model.AdPositionMappingInput) (*model.AdPositionMapping, error) {
	return restCall[*model.AdPositionMapping](ctx, r.Proxy, "admanagement", rest.Request{Method: "POST", Path: "/position-mappings", Body: input}, "")
}

func (r *mutationResolver) AdUpdatePositionMapping(ctx context.Context, id int, input model.AdPositionMappingInput) (*model.AdPositionMapping, error) {
	return restCall[*model.AdPositionMapping](ctx, r.Proxy, "admanagement", rest.Request{Method: "PUT", Path: rest.P("/position-mappings/%v", id), Body: input}, "")
}

func (r *mutationResolver) AdDeletePositionMapping(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "admanagement", rest.Request{Method: "DELETE", Path: rest.P("/position-mappings/%v", id)}, "")
}

func (r *mutationResolver) AdCreateTargetingRule(ctx context.Context, input model.AdTargetingRuleInput) (*model.AdTargetingRule, error) {
	return restCall[*model.AdTargetingRule](ctx, r.Proxy, "admanagement", rest.Request{Method: "POST", Path: "/targeting-rules", Body: input}, "")
}

func (r *mutationResolver) AdUpdateTargetingRule(ctx context.Context, id int, input model.AdTargetingRuleInput) (*model.AdTargetingRule, error) {
	return restCall[*model.AdTargetingRule](ctx, r.Proxy, "admanagement", rest.Request{Method: "PUT", Path: rest.P("/targeting-rules/%v", id), Body: input}, "")
}

func (r *mutationResolver) AdDeleteTargetingRule(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "admanagement", rest.Request{Method: "DELETE", Path: rest.P("/targeting-rules/%v", id)}, "")
}

func (r *queryResolver) AdCampaigns(ctx context.Context) ([]*model.AdCampaign, error) {
	return restCall[[]*model.AdCampaign](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/campaigns"}, "")
}

func (r *queryResolver) AdCampaign(ctx context.Context, id int) (*model.AdCampaign, error) {
	return restCall[*model.AdCampaign](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/campaigns/%v", id)}, "")
}

func (r *queryResolver) AdCampaignAds(ctx context.Context, id int) ([]*model.AdAd, error) {
	return restCall[[]*model.AdAd](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/campaigns/%v/ads", id)}, "")
}

func (r *queryResolver) AdCampaignPerformance(ctx context.Context, id int, from *string, to *string) (*model.AdCampaignPerformance, error) {
	return restCall[*model.AdCampaignPerformance](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/campaigns/%v/performance", id), Query: restQuery("from", from, "to", to)}, "")
}

func (r *queryResolver) AdCategories(ctx context.Context) ([]*model.AdCategory, error) {
	return restCall[[]*model.AdCategory](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/categories"}, "")
}

func (r *queryResolver) AdCategory(ctx context.Context, id int) (*model.AdCategory, error) {
	return restCall[*model.AdCategory](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/categories/%v", id)}, "")
}

func (r *queryResolver) AdPositions(ctx context.Context) ([]*model.AdPosition, error) {
	return restCall[[]*model.AdPosition](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/positions"}, "")
}

func (r *queryResolver) AdPosition(ctx context.Context, id int) (*model.AdPosition, error) {
	return restCall[*model.AdPosition](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/positions/%v", id)}, "")
}

func (r *queryResolver) AdAds(ctx context.Context) ([]*model.AdAd, error) {
	return restCall[[]*model.AdAd](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/ads"}, "")
}

func (r *queryResolver) AdAd(ctx context.Context, id int) (*model.AdAd, error) {
	return restCall[*model.AdAd](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/ads/%v", id)}, "")
}

func (r *queryResolver) AdAdImpressions(ctx context.Context, id int) ([]*model.AdImpression, error) {
	return restCall[[]*model.AdImpression](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/ads/%v/impressions", id)}, "")
}

func (r *queryResolver) AdAdClicks(ctx context.Context, id int) ([]*model.AdClick, error) {
	return restCall[[]*model.AdClick](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/ads/%v/clicks", id)}, "")
}

func (r *queryResolver) AdPositionMappings(ctx context.Context) ([]*model.AdPositionMapping, error) {
	return restCall[[]*model.AdPositionMapping](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/position-mappings"}, "")
}

func (r *queryResolver) AdPositionMapping(ctx context.Context, id int) (*model.AdPositionMapping, error) {
	return restCall[*model.AdPositionMapping](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/position-mappings/%v", id)}, "")
}

func (r *queryResolver) AdImpressions(ctx context.Context) ([]*model.AdImpression, error) {
	return restCall[[]*model.AdImpression](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/impressions"}, "")
}

func (r *queryResolver) AdImpression(ctx context.Context, id int) (*model.AdImpression, error) {
	return restCall[*model.AdImpression](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/impressions/%v", id)}, "")
}

func (r *queryResolver) AdClicks(ctx context.Context) ([]*model.AdClick, error) {
	return restCall[[]*model.AdClick](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/clicks"}, "")
}

func (r *queryResolver) AdClick(ctx context.Context, id int) (*model.AdClick, error) {
	return restCall[*model.AdClick](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/clicks/%v", id)}, "")
}

func (r *queryResolver) AdTargetingRules(ctx context.Context) ([]*model.AdTargetingRule, error) {
	return restCall[[]*model.AdTargetingRule](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/targeting-rules"}, "")
}

func (r *queryResolver) AdTargetingRule(ctx context.Context, id int) (*model.AdTargetingRule, error) {
	return restCall[*model.AdTargetingRule](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/targeting-rules/%v", id)}, "")
}

func (r *queryResolver) AdPerformanceSummaries(ctx context.Context) ([]*model.AdPerformanceSummary, error) {
	return restCall[[]*model.AdPerformanceSummary](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/performance-summaries"}, "")
}

func (r *queryResolver) AdPerformanceSummary(ctx context.Context, id int) (*model.AdPerformanceSummary, error) {
	return restCall[*model.AdPerformanceSummary](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/performance-summaries/%v", id)}, "")
}

func (r *queryResolver) AdServe(ctx context.Context, positionID int, sessionID *string, country *string, device *string, gender *string, age *int, referrerURL *string, pageURL *string) (*model.AdServedAd, error) {
	return restCall[*model.AdServedAd](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: rest.P("/positions/%v/serve", positionID), Query: restQuery("session_id", sessionID, "country", country, "device", device, "gender", gender, "age", age, "referrer_url", referrerURL, "page_url", pageURL)}, "")
}

func (r *queryResolver) AdTrackClick(ctx context.Context, adID int, impressionID *int, sessionID *string, referrerURL *string) (*model.AdClickResult, error) {
	return restCall[*model.AdClickResult](ctx, r.Proxy, "admanagement", rest.Request{Method: "GET", Path: "/clicks/track", Query: restQuery("ad_id", adID, "impression_id", impressionID, "session_id", sessionID, "referrer_url", referrerURL)}, "")
}
