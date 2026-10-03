package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) GymQuery() generated.GymQueryResolver { return &gymQueryResolver{r} }

type gymQueryResolver struct{ *Resolver }

func (r *gymQueryResolver) Classes(ctx context.Context, obj *model.GymQuery) ([]*model.GymClass, error) {
	return r.Clients.IdentifiedService.Classes(ctx)
}

func (r *gymQueryResolver) Class(ctx context.Context, obj *model.GymQuery, classID string) (*model.GymClass, error) {
	return r.Clients.IdentifiedService.Class(ctx, classID)
}

func (r *gymQueryResolver) GymPassVerification(ctx context.Context, obj *model.GymQuery, gymPassID string) (*model.GymPass, error) {
	return r.Clients.IdentifiedService.GymPassVerification(ctx, gymPassID)
}
