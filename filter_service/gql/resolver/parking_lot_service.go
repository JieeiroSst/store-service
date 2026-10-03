package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ParkingQuery() generated.ParkingQueryResolver { return &parkingQueryResolver{r} }

type parkingQueryResolver struct{ *Resolver }

func (r *parkingQueryResolver) AvailableSpots(ctx context.Context, obj *model.ParkingQuery) ([]*model.ParkingSpotAvailability, error) {
	return r.Clients.ParkingLotService.AvailableSpots(ctx)
}

func (r *parkingQueryResolver) Rates(ctx context.Context, obj *model.ParkingQuery) ([]*model.ParkingRate, error) {
	return r.Clients.ParkingLotService.Rates(ctx)
}

func (r *parkingQueryResolver) Ticket(ctx context.Context, obj *model.ParkingQuery, id string) (*model.ParkingTicket, error) {
	return r.Clients.ParkingLotService.Ticket(ctx, id)
}

func (r *parkingQueryResolver) History(ctx context.Context, obj *model.ParkingQuery, plate string, limit *int, offset *int) ([]*model.ParkingTicket, error) {
	return r.Clients.ParkingLotService.History(ctx, plate, limit, offset)
}
