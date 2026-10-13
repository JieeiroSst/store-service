package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) BonusRecordReward(ctx context.Context, input model.BonusRecordRewardInput) (*model.BonusRecordResult, error) {
	return restCall[*model.BonusRecordResult](ctx, r.Proxy, "bonuslink-service", rest.Request{Method: "POST", Path: "/api/v1/bonus/rewards", Body: input}, "data")
}

func (r *queryResolver) BonusUserRewards(ctx context.Context, userID string, limit *int, offset *int) ([]*model.BonusReward, error) {
	return restCall[[]*model.BonusReward](ctx, r.Proxy, "bonuslink-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/bonus/users/%v/rewards", userID), Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) BonusUserBalances(ctx context.Context, userID string) ([]*model.BonusBalance, error) {
	return restCall[[]*model.BonusBalance](ctx, r.Proxy, "bonuslink-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/bonus/users/%v/balances", userID)}, "data")
}

func (r *queryResolver) MyBonusRewards(ctx context.Context, limit *int, offset *int) ([]*model.BonusReward, error) {
	return restCall[[]*model.BonusReward](ctx, r.Proxy, "bonuslink-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/bonus/users/%v/rewards", userIDOf(ctx)), Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) MyBonusBalances(ctx context.Context) ([]*model.BonusBalance, error) {
	return restCall[[]*model.BonusBalance](ctx, r.Proxy, "bonuslink-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/bonus/users/%v/balances", userIDOf(ctx))}, "data")
}
