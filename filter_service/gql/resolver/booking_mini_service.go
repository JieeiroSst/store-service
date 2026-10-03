package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) BookingMiniQuery() generated.BookingMiniQueryResolver {
	return &bookingMiniQueryResolver{r}
}

type bookingMiniQueryResolver struct{ *Resolver }

func (r *bookingMiniQueryResolver) AvailableRooms(ctx context.Context, obj *model.BookingMiniQuery, fromDate string, toDate string) ([]*model.BookingMiniRoom, error) {
	return r.Clients.BookingMiniService.AvailableRooms(ctx, fromDate, toDate)
}

func (r *bookingMiniQueryResolver) Booking(ctx context.Context, obj *model.BookingMiniQuery, id int) (*model.BookingMiniBooking, error) {
	return r.Clients.BookingMiniService.Booking(ctx, id)
}

func (r *bookingMiniQueryResolver) Passenger(ctx context.Context, obj *model.BookingMiniQuery, id int) (*model.BookingMiniPassenger, error) {
	return r.Clients.BookingMiniService.Passenger(ctx, id)
}
