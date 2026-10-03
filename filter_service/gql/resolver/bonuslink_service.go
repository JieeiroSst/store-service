package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) BonusQuery() generated.BonusQueryResolver { return &bonusQueryResolver{r} }

type bonusQueryResolver struct{ *Resolver }

func (r *bonusQueryResolver) UserRewards(ctx context.Context, obj *model.BonusQuery, userID string, limit *int, offset *int) (*model.BonusRewardList, error) {
	return r.Clients.BonuslinkService.UserRewards(ctx, userID, limit, offset)
}

func (r *bonusQueryResolver) UserBalances(ctx context.Context, obj *model.BonusQuery, userID string) ([]*model.BonusBalance, error) {
	return r.Clients.BonuslinkService.UserBalances(ctx, userID)
}
