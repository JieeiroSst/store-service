package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) DoordashQuery() generated.DoordashQueryResolver { return &doordashQueryResolver{r} }

type doordashQueryResolver struct{ *Resolver }

func (r *doordashQueryResolver) Orders(ctx context.Context, obj *model.DoordashQuery, customerID *string, restaurantID *string) ([]*model.DoordashOrder, error) {
	return r.Clients.DoordashService.Orders(ctx, customerID, restaurantID)
}

func (r *doordashQueryResolver) Order(ctx context.Context, obj *model.DoordashQuery, id int) (*model.DoordashOrder, error) {
	return r.Clients.DoordashService.Order(ctx, id)
}

func (r *doordashQueryResolver) OrderTracking(ctx context.Context, obj *model.DoordashQuery, id int) ([]*model.DoordashOrderTracking, error) {
	return r.Clients.DoordashService.OrderTracking(ctx, id)
}

func (r *doordashQueryResolver) DriverAssignments(ctx context.Context, obj *model.DoordashQuery, driverID string) ([]*model.DoordashDriverAssignment, error) {
	return r.Clients.DoordashService.DriverAssignments(ctx, driverID)
}
