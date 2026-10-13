package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ParkingCheckIn(ctx context.Context, licensePlate string, vehicleType string, entryGateID *string) (*model.ParkingTicket, error) {
	return restCall[*model.ParkingTicket](ctx, r.Proxy, "parking-lot-service", rest.Request{Method: "POST", Path: "/api/v1/check-in", Body: map[string]any{"license_plate": licensePlate, "vehicle_type": vehicleType, "entry_gate_id": entryGateID}}, "")
}

func (r *mutationResolver) ParkingCheckOut(ctx context.Context, ticketID string, method string, exitGateID *string) (*model.ParkingCheckOut, error) {
	return restCall[*model.ParkingCheckOut](ctx, r.Proxy, "parking-lot-service", rest.Request{Method: "POST", Path: "/api/v1/check-out", Body: map[string]any{"ticket_id": ticketID, "method": method, "exit_gate_id": exitGateID}}, "")
}

func (r *queryResolver) ParkingAvailableSpots(ctx context.Context) ([]*model.ParkingSpotAvailability, error) {
	return restCall[[]*model.ParkingSpotAvailability](ctx, r.Proxy, "parking-lot-service", rest.Request{Method: "GET", Path: "/api/v1/spots/available"}, "")
}

func (r *queryResolver) ParkingRates(ctx context.Context) ([]*model.ParkingRate, error) {
	return restCall[[]*model.ParkingRate](ctx, r.Proxy, "parking-lot-service", rest.Request{Method: "GET", Path: "/api/v1/rates"}, "")
}

func (r *queryResolver) ParkingTicket(ctx context.Context, id string) (*model.ParkingTicket, error) {
	return restCall[*model.ParkingTicket](ctx, r.Proxy, "parking-lot-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/tickets/%v", id)}, "")
}

func (r *queryResolver) ParkingHistory(ctx context.Context, plate string, limit *int, offset *int) ([]*model.ParkingTicket, error) {
	return restCall[[]*model.ParkingTicket](ctx, r.Proxy, "parking-lot-service", rest.Request{Method: "GET", Path: "/api/v1/history", Query: restQuery("plate", plate, "limit", limit, "offset", offset)}, "")
}
