package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) PointQuery() generated.PointQueryResolver { return &pointQueryResolver{r} }

type pointQueryResolver struct{ *Resolver }

func (r *pointQueryResolver) RewardPoints(ctx context.Context, obj *model.PointQuery, perPage *int, sortOrder *string, cursor *string) (*model.PointRewardPointPage, error) {
	return r.Clients.PointService.RewardPoints(ctx, perPage, sortOrder, cursor)
}

func (r *pointQueryResolver) RewardPoint(ctx context.Context, obj *model.PointQuery, id string) (*model.PointRewardPoint, error) {
	return r.Clients.PointService.RewardPoint(ctx, id)
}

func (r *pointQueryResolver) RewardDiscounts(ctx context.Context, obj *model.PointQuery, perPage *int, sortOrder *string, cursor *string) (*model.PointRewardDiscountPage, error) {
	return r.Clients.PointService.RewardDiscounts(ctx, perPage, sortOrder, cursor)
}

func (r *pointQueryResolver) RewardDiscount(ctx context.Context, obj *model.PointQuery, id string) (*model.PointRewardDiscount, error) {
	return r.Clients.PointService.RewardDiscount(ctx, id)
}

func (r *pointQueryResolver) ConvertedRewardPoints(ctx context.Context, obj *model.PointQuery, perPage *int, sortOrder *string, cursor *string) (*model.PointConvertedRewardPointPage, error) {
	return r.Clients.PointService.ConvertedRewardPoints(ctx, perPage, sortOrder, cursor)
}

func (r *pointQueryResolver) ConvertedRewardPoint(ctx context.Context, obj *model.PointQuery, id string) (*model.PointConvertedRewardPoint, error) {
	return r.Clients.PointService.ConvertedRewardPoint(ctx, id)
}
