package port

import (
	"context"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
)

type ParkingUsecase interface {
	CheckIn(ctx context.Context, plate string, vehicleType model.VehicleType, entryGateID *string) (*model.Ticket, error)
	CheckOut(ctx context.Context, ticketID string, method model.PaymentMethod, exitGateID *string) (*model.Ticket, *model.Payment, error)
	GetTicket(ctx context.Context, ticketID string) (*model.Ticket, error)
	ListHistory(ctx context.Context, plate string, limit, offset int) ([]model.Ticket, error)
	AvailableSpots(ctx context.Context) ([]model.SpotAvailability, error)
}

type RateUsecase interface {
	ListRates(ctx context.Context) ([]model.Rate, error)
}
