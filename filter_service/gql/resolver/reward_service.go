package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RewardQuery() generated.RewardQueryResolver { return &rewardQueryResolver{r} }

type rewardQueryResolver struct{ *Resolver }

func (r *rewardQueryResolver) Rewards(ctx context.Context, obj *model.RewardQuery, name *string, page *int, pageSize *int) (*model.RewardPage, error) {
	return r.Clients.RewardService.Rewards(ctx, name, page, pageSize)
}

func (r *rewardQueryResolver) Reward(ctx context.Context, obj *model.RewardQuery, rewardID string) (*model.RewardPage, error) {
	return r.Clients.RewardService.Reward(ctx, rewardID)
}
