package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) BookingMiniCreateBooking(ctx context.Context, input model.BookingMiniBookingInput) (*model.BookingMiniBooking, error) {
	return restCall[*model.BookingMiniBooking](ctx, r.Proxy, "booking-mini-service", rest.Request{Method: "POST", Path: "/api/bookings", Body: input}, "")
}

func (r *mutationResolver) BookingMiniCancelBooking(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "booking-mini-service", rest.Request{Method: "DELETE", Path: rest.P("/api/bookings/%v", id)}, "")
}

func (r *mutationResolver) BookingMiniRegisterPassenger(ctx context.Context, name string) (*model.BookingMiniPassenger, error) {
	return restCall[*model.BookingMiniPassenger](ctx, r.Proxy, "booking-mini-service", rest.Request{Method: "POST", Path: "/api/passengers", Body: map[string]any{"name": name}}, "")
}

func (r *queryResolver) BookingMiniAvailableRooms(ctx context.Context, fromDate string, toDate string) ([]*model.BookingMiniRoom, error) {
	return restCall[[]*model.BookingMiniRoom](ctx, r.Proxy, "booking-mini-service", rest.Request{Method: "GET", Path: "/api/rooms", Query: restQuery("from_date", fromDate, "to_date", toDate)}, "")
}

func (r *queryResolver) BookingMiniBooking(ctx context.Context, id int) (*model.BookingMiniBooking, error) {
	return restCall[*model.BookingMiniBooking](ctx, r.Proxy, "booking-mini-service", rest.Request{Method: "GET", Path: rest.P("/api/bookings/%v", id)}, "")
}

func (r *queryResolver) BookingMiniPassenger(ctx context.Context, id int) (*model.BookingMiniPassenger, error) {
	return restCall[*model.BookingMiniPassenger](ctx, r.Proxy, "booking-mini-service", rest.Request{Method: "GET", Path: rest.P("/api/passengers/%v", id)}, "")
}
