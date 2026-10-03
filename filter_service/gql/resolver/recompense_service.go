package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RecompQuery() generated.RecompQueryResolver { return &recompQueryResolver{r} }

type recompQueryResolver struct{ *Resolver }

func (r *recompQueryResolver) Recompenses(ctx context.Context, obj *model.RecompQuery, page *int, limit *int) ([]*model.RecompRecompense, error) {
	return r.Clients.RecompenseService.Recompenses(ctx, page, limit)
}

func (r *recompQueryResolver) Recompense(ctx context.Context, obj *model.RecompQuery, id string) (*model.RecompRecompense, error) {
	return r.Clients.RecompenseService.Recompense(ctx, id)
}

func (r *recompQueryResolver) MemberRecompenses(ctx context.Context, obj *model.RecompQuery, memberID int, page *int, limit *int) ([]*model.RecompRecompense, error) {
	return r.Clients.RecompenseService.MemberRecompenses(ctx, memberID, page, limit)
}

func (r *recompQueryResolver) MemberPoints(ctx context.Context, obj *model.RecompQuery, memberID int) (*model.RecompMemberStatus, error) {
	return r.Clients.RecompenseService.MemberPoints(ctx, memberID)
}

func (r *recompQueryResolver) BestRecompense(ctx context.Context, obj *model.RecompQuery, memberID int, amount float64) (*model.RecompRedemption, error) {
	return r.Clients.RecompenseService.BestRecompense(ctx, memberID, amount)
}
