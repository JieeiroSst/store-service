package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) LotteryQuery() generated.LotteryQueryResolver { return &lotteryQueryResolver{r} }

type lotteryQueryResolver struct{ *Resolver }

func (r *lotteryQueryResolver) Results(ctx context.Context, obj *model.LotteryQuery, date string, province string) (*model.LotteryResult, error) {
	return r.Clients.LotteryService.Results(ctx, date, province)
}
