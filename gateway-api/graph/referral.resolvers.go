package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) GenerateMyReferralLink(ctx context.Context, channel *string, platform *string, utmSource *string, utmMedium *string, utmCampaign *string, utmTerm *string, utmContent *string) (*model.ReferralReferralLink, error) {
	return restCall[*model.ReferralReferralLink](ctx, r.Proxy, "referral-service", rest.Request{Method: "POST", Path: "/api/v1/referral/generate", Body: withField(map[string]any{"channel": channel, "platform": platform, "utm_source": utmSource, "utm_medium": utmMedium, "utm_campaign": utmCampaign, "utm_term": utmTerm, "utm_content": utmContent}, "owner_user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) ReferralTrackEvent(ctx context.Context, refCode string, eventType string, platform *string, channel *string, deviceID *string, firebaseInstanceID *string, utmSource *string, utmMedium *string, utmCampaign *string, utmTerm *string, utmContent *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "referral-service", rest.Request{Method: "POST", Path: "/api/v1/referral/event", Body: map[string]any{"ref_code": refCode, "event_type": eventType, "platform": platform, "channel": channel, "device_id": deviceID, "firebase_instance_id": firebaseInstanceID, "utm_source": utmSource, "utm_medium": utmMedium, "utm_campaign": utmCampaign, "utm_term": utmTerm, "utm_content": utmContent}}, "")
}

func (r *mutationResolver) ReferralConfirmInstall(ctx context.Context, refCode string, clickID *string, platform *string, deviceID *string, firebaseInstanceID *string, utmSource *string, utmMedium *string, utmCampaign *string, utmTerm *string, utmContent *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "referral-service", rest.Request{Method: "POST", Path: "/api/v1/referral/confirm-install", Body: withField(map[string]any{"ref_code": refCode, "click_id": clickID, "platform": platform, "device_id": deviceID, "firebase_instance_id": firebaseInstanceID, "utm_source": utmSource, "utm_medium": utmMedium, "utm_campaign": utmCampaign, "utm_term": utmTerm, "utm_content": utmContent}, "new_user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) ReferralActivate(ctx context.Context, refCode string, clickID *string, platform *string, deviceID *string, firebaseInstanceID *string, utmSource *string, utmMedium *string, utmCampaign *string, utmTerm *string, utmContent *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "referral-service", rest.Request{Method: "POST", Path: "/api/v1/referral/activate", Body: withField(map[string]any{"ref_code": refCode, "click_id": clickID, "platform": platform, "device_id": deviceID, "firebase_instance_id": firebaseInstanceID, "utm_source": utmSource, "utm_medium": utmMedium, "utm_campaign": utmCampaign, "utm_term": utmTerm, "utm_content": utmContent}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) ReferralCreateRewardProgram(ctx context.Context, name string, activate *bool, tiers []*model.ReferralRewardTierInput) (*model.ReferralRewardProgram, error) {
	return restCall[*model.ReferralRewardProgram](ctx, r.Proxy, "referral-service", rest.Request{Method: "POST", Path: "/api/v1/programs", Body: map[string]any{"name": name, "activate": activate, "tiers": tiers}}, "")
}

func (r *mutationResolver) ReferralSetRewardProgramStatus(ctx context.Context, programID string, active bool) (*model.ReferralRewardProgram, error) {
	return restCall[*model.ReferralRewardProgram](ctx, r.Proxy, "referral-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/programs/%v/status", programID), Body: map[string]any{"active": active}}, "")
}

func (r *queryResolver) ReferralLink(ctx context.Context, refCode string) (*model.ReferralReferralLink, error) {
	return restCall[*model.ReferralReferralLink](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referral/link/%v", refCode)}, "data")
}

func (r *queryResolver) ReferralUserLinks(ctx context.Context, userID string, limit *int, cursor *string) (*model.ReferralLinkList, error) {
	return restCall[*model.ReferralLinkList](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referral/user/%v/links", userID), Query: restQuery("limit", limit, "cursor", cursor)}, "")
}

func (r *queryResolver) ReferralStatus(ctx context.Context, refCode string) (*model.ReferralReferralStatusResponse, error) {
	return restCall[*model.ReferralReferralStatusResponse](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: "/api/v1/referral/status", Query: restQuery("ref_code", refCode)}, "data")
}

func (r *queryResolver) ReferralUserStats(ctx context.Context, userID string) (*model.ReferralUserReferralStats, error) {
	return restCall[*model.ReferralUserReferralStats](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referral/user/%v/stats", userID)}, "data")
}

func (r *queryResolver) ReferralShareTargets(ctx context.Context, refCode string) ([]*model.ReferralShareTarget, error) {
	return restCall[[]*model.ReferralShareTarget](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referral/share/%v", refCode)}, "data")
}

func (r *queryResolver) ReferralActiveRewardProgram(ctx context.Context) (*model.ReferralRewardProgram, error) {
	return restCall[*model.ReferralRewardProgram](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: "/api/v1/programs/active"}, "data")
}

func (r *queryResolver) MyReferralLinks(ctx context.Context, limit *int, cursor *string) (*model.ReferralLinkList, error) {
	return restCall[*model.ReferralLinkList](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referral/user/%v/links", userIDOf(ctx)), Query: restQuery("limit", limit, "cursor", cursor)}, "")
}

func (r *queryResolver) MyReferralStats(ctx context.Context) (*model.ReferralUserReferralStats, error) {
	return restCall[*model.ReferralUserReferralStats](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referral/user/%v/stats", userIDOf(ctx))}, "")
}

func (r *queryResolver) ReferralUserAttribution(ctx context.Context, userID string) (*model.ReferralAttribution, error) {
	return restCall[*model.ReferralAttribution](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referral/user/%v/attribution", userID)}, "data")
}

func (r *queryResolver) MyReferralAttribution(ctx context.Context) (*model.ReferralAttribution, error) {
	return restCall[*model.ReferralAttribution](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/referral/user/%v/attribution", userIDOf(ctx))}, "data")
}

func (r *queryResolver) ReferralUTMReport(ctx context.Context, ownerUserID *string, utmSource *string, utmMedium *string, utmCampaign *string, from *int, to *int) ([]*model.ReferralUTMPerformance, error) {
	return restCall[[]*model.ReferralUTMPerformance](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: "/api/v1/referral/utm/report", Query: restQuery("owner_user_id", ownerUserID, "utm_source", utmSource, "utm_medium", utmMedium, "utm_campaign", utmCampaign, "from", from, "to", to)}, "data")
}

func (r *queryResolver) MyReferralUTMReport(ctx context.Context, utmSource *string, utmMedium *string, utmCampaign *string, from *int, to *int) ([]*model.ReferralUTMPerformance, error) {
	return restCall[[]*model.ReferralUTMPerformance](ctx, r.Proxy, "referral-service", rest.Request{Method: "GET", Path: "/api/v1/referral/utm/report", Query: restQuery("utm_source", utmSource, "utm_medium", utmMedium, "utm_campaign", utmCampaign, "from", from, "to", to, "owner_user_id", userIDOf(ctx))}, "data")
}
