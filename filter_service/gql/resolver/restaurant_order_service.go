package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ROrderQuery() generated.ROrderQueryResolver { return &rOrderQueryResolver{r} }

type rOrderQueryResolver struct{ *Resolver }

func (r *rOrderQueryResolver) Order(ctx context.Context, obj *model.ROrderQuery, id int) (*model.ROrderOrder, error) {
	return r.Clients.RestaurantOrderService.Order(ctx, id)
}

func (r *rOrderQueryResolver) Orders(ctx context.Context, obj *model.ROrderQuery, limit *int, page *int, sort *string) (*model.ROrderOrderPage, error) {
	return r.Clients.RestaurantOrderService.Orders(ctx, limit, page, sort)
}

func (r *rOrderQueryResolver) Reservations(ctx context.Context, obj *model.ROrderQuery, limit *int, page *int, sort *string) (*model.ROrderReservationPage, error) {
	return r.Clients.RestaurantOrderService.Reservations(ctx, limit, page, sort)
}

func (r *rOrderQueryResolver) Reservation(ctx context.Context, obj *model.ROrderQuery, id int) (*model.ROrderReservation, error) {
	return r.Clients.RestaurantOrderService.Reservation(ctx, id)
}
