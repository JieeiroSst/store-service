package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ArrangeQuery() generated.ArrangeQueryResolver { return &arrangeQueryResolver{r} }

type arrangeQueryResolver struct{ *Resolver }

func (r *arrangeQueryResolver) Shifts(ctx context.Context, obj *model.ArrangeQuery, storeID int) ([]*model.ArrangeShift, error) {
	return r.Clients.ArrangeService.Shifts(ctx, storeID)
}

func (r *arrangeQueryResolver) AvailableSpaces(ctx context.Context, obj *model.ArrangeQuery, storeID int, minCapacity *int) ([]*model.ArrangeSpace, error) {
	return r.Clients.ArrangeService.AvailableSpaces(ctx, storeID, minCapacity)
}

func (r *arrangeQueryResolver) SpaceAssignments(ctx context.Context, obj *model.ArrangeQuery, storeID int) ([]*model.ArrangeSpaceAssignment, error) {
	return r.Clients.ArrangeService.SpaceAssignments(ctx, storeID)
}

func (r *arrangeQueryResolver) DispatchAssignments(ctx context.Context, obj *model.ArrangeQuery, driverID int) ([]*model.ArrangeDispatchAssignment, error) {
	return r.Clients.ArrangeService.DispatchAssignments(ctx, driverID)
}
