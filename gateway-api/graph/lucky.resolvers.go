package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) LuckyPlay(ctx context.Context, game string, choice *int, idempotencyKey *string) (*model.LuckyPlayResult, error) {
	return restCall[*model.LuckyPlayResult](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/lucky/games/%v/me/plays", game), Body: map[string]any{"choice": choice, "idempotency_key": idempotencyKey}, Header: restHeader("Idempotency-Key", idempotencyKey)}, "data")
}

func (r *mutationResolver) LuckyCheckIn(ctx context.Context) (*model.LuckyCheckinResult, error) {
	return restCall[*model.LuckyCheckinResult](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: "/api/v1/lucky/checkin/me", Body: map[string]any{}}, "data")
}

func (r *mutationResolver) LuckyGrantPlays(ctx context.Context, userID string, event string, referenceID string) (*model.LuckyGrantResult, error) {
	return restCall[*model.LuckyGrantResult](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: "/api/v1/lucky/credits", Body: map[string]any{"user_id": userID, "event": event, "reference_id": referenceID}}, "data")
}

func (r *mutationResolver) LuckyCampaignCreate(ctx context.Context, input model.LuckyCampaignInput) (*model.LuckyCampaign, error) {
	return restCall[*model.LuckyCampaign](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: "/api/v1/lucky/campaigns", Body: input}, "data")
}

func (r *mutationResolver) LuckyCampaignUpdate(ctx context.Context, id int, input model.LuckyCampaignInput) (*model.LuckyCampaign, error) {
	return restCall[*model.LuckyCampaign](ctx, r.Proxy, "lucky-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/lucky/campaigns/%v", id), Body: input}, "data")
}

func (r *mutationResolver) LuckyCampaignDelete(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "lucky-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/lucky/campaigns/%v", id)}, "")
}

func (r *mutationResolver) LuckyCampaignActivate(ctx context.Context, id int) (*model.LuckyCampaign, error) {
	return restCall[*model.LuckyCampaign](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/lucky/campaigns/%v/activate", id), Body: map[string]any{}}, "data")
}

func (r *mutationResolver) LuckyFraudBlock(ctx context.Context, userID string, reason string) (*model.LuckyBlockResult, error) {
	return restCall[*model.LuckyBlockResult](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: "/api/v1/lucky/fraud/blocks", Body: map[string]any{"user_id": userID, "reason": reason}}, "data")
}

func (r *mutationResolver) LuckyFraudUnblock(ctx context.Context, userID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "lucky-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/lucky/fraud/blocks/%v", userID)}, "")
}

func (r *mutationResolver) LuckyFraudApproveReward(ctx context.Context, id int) (*model.LuckyReviewReward, error) {
	return restCall[*model.LuckyReviewReward](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/lucky/fraud/rewards/%v/approve", id), Body: map[string]any{}}, "data")
}

func (r *mutationResolver) LuckyFraudRejectReward(ctx context.Context, id int) (*model.LuckyReviewReward, error) {
	return restCall[*model.LuckyReviewReward](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/lucky/fraud/rewards/%v/reject", id), Body: map[string]any{}}, "data")
}

func (r *mutationResolver) LuckyCampaignPause(ctx context.Context, id int) (*model.LuckyCampaign, error) {
	return restCall[*model.LuckyCampaign](ctx, r.Proxy, "lucky-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/lucky/campaigns/%v/pause", id), Body: map[string]any{}}, "data")
}

func (r *queryResolver) LuckyGames(ctx context.Context) ([]*model.LuckyCampaign, error) {
	return restCall[[]*model.LuckyCampaign](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: "/api/v1/lucky/games"}, "data")
}

func (r *queryResolver) LuckyGame(ctx context.Context, game string) (*model.LuckyCampaign, error) {
	return restCall[*model.LuckyCampaign](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/lucky/games/%v", game)}, "data")
}

func (r *queryResolver) LuckyWinners(ctx context.Context, game string, limit *int) ([]*model.LuckyWinner, error) {
	return restCall[[]*model.LuckyWinner](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/lucky/games/%v/winners", game), Query: restQuery("limit", limit)}, "data")
}

func (r *queryResolver) LuckyMyStatus(ctx context.Context, game string) (*model.LuckyPlayerStatus, error) {
	return restCall[*model.LuckyPlayerStatus](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/lucky/games/%v/me/status", game)}, "data")
}

func (r *queryResolver) LuckyMyPlays(ctx context.Context, gameType *string, campaignID *int, limit *int, offset *int) (*model.LuckyPlayPage, error) {
	return restCall[*model.LuckyPlayPage](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: "/api/v1/lucky/me/plays", Query: restQuery("game_type", gameType, "campaign_id", campaignID, "limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) LuckyCheckinStatus(ctx context.Context) (*model.LuckyCheckinStatus, error) {
	return restCall[*model.LuckyCheckinStatus](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: "/api/v1/lucky/checkin/me"}, "data")
}

func (r *queryResolver) LuckyCheckinHistory(ctx context.Context, limit *int, offset *int) (*model.LuckyCheckinPage, error) {
	return restCall[*model.LuckyCheckinPage](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: "/api/v1/lucky/checkin/me/history", Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) LuckyCampaigns(ctx context.Context, gameType *string, status *string, limit *int, offset *int) (*model.LuckyCampaignPage, error) {
	return restCall[*model.LuckyCampaignPage](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: "/api/v1/lucky/campaigns", Query: restQuery("game_type", gameType, "status", status, "limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) LuckyCampaign(ctx context.Context, id int) (*model.LuckyCampaign, error) {
	return restCall[*model.LuckyCampaign](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/lucky/campaigns/%v", id)}, "data")
}

func (r *queryResolver) LuckyFraudBlocks(ctx context.Context, limit *int, offset *int) (*model.LuckyBlockPage, error) {
	return restCall[*model.LuckyBlockPage](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: "/api/v1/lucky/fraud/blocks", Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) LuckyFraudRewards(ctx context.Context, status *string, userID *string, limit *int, offset *int) (*model.LuckyReviewRewardPage, error) {
	return restCall[*model.LuckyReviewRewardPage](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: "/api/v1/lucky/fraud/rewards", Query: restQuery("status", status, "user_id", userID, "limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) LuckyFraudPlays(ctx context.Context, userID *string, deviceID *string, ip *string, limit *int, offset *int) (*model.LuckyPlayPage, error) {
	return restCall[*model.LuckyPlayPage](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: "/api/v1/lucky/fraud/plays", Query: restQuery("user_id", userID, "device_id", deviceID, "ip", ip, "limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) LuckyCampaignPlays(ctx context.Context, id int, userID *string, limit *int, offset *int) (*model.LuckyPlayPage, error) {
	return restCall[*model.LuckyPlayPage](ctx, r.Proxy, "lucky-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/lucky/campaigns/%v/plays", id), Query: restQuery("user_id", userID, "limit", limit, "offset", offset)}, "data")
}
