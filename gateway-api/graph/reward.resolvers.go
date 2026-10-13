package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RewardCreateReward(ctx context.Context, input model.RewardRewardInput) (*model.RewardReward, error) {
	return restCall[*model.RewardReward](ctx, r.Proxy, "reward-service", rest.Request{Method: "POST", Path: "/reward", Body: input}, "")
}

func (r *mutationResolver) RewardUpdateReward(ctx context.Context, rewardID string, input model.RewardRewardInput) (*model.RewardReward, error) {
	return restCall[*model.RewardReward](ctx, r.Proxy, "reward-service", rest.Request{Method: "PUT", Path: rest.P("/reward/%v", rewardID), Body: input}, "")
}

func (r *mutationResolver) RewardDeleteReward(ctx context.Context, rewardID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "reward-service", rest.Request{Method: "DELETE", Path: rest.P("/reward/%v", rewardID)}, "")
}

func (r *queryResolver) Rewards(ctx context.Context, name *string, page *int, pageSize *int) (*model.RewardPage, error) {
	return restCall[*model.RewardPage](ctx, r.Proxy, "reward-service", rest.Request{Method: "GET", Path: "/reward", Query: restQuery("name", name, "page", page, "page_size", pageSize)}, "")
}

func (r *queryResolver) Reward(ctx context.Context, rewardID string) (*model.RewardPage, error) {
	return restCall[*model.RewardPage](ctx, r.Proxy, "reward-service", rest.Request{Method: "GET", Path: rest.P("/reward/%v", rewardID)}, "")
}
