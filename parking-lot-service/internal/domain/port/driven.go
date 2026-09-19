package port

import (
	"context"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
)

type VehicleRepository interface {
	GetOrCreate(ctx context.Context, plate string, vehicleType model.VehicleType) (*model.Vehicle, error)
	Get(ctx context.Context, plate string) (*model.Vehicle, error)
}

type ParkingSpotRepository interface {
	FindAndReserveAvailable(ctx context.Context, types []model.SpotType) (*model.ParkingSpot, error)
	Release(ctx context.Context, spotID string) error
	CountByType(ctx context.Context) ([]model.SpotAvailability, error)
}

type TicketRepository interface {
	Create(ctx context.Context, ticket *model.Ticket) error
	GetByID(ctx context.Context, id string) (*model.Ticket, error)
	Update(ctx context.Context, ticket *model.Ticket) error
	ListByPlate(ctx context.Context, plate string, limit, offset int) ([]model.Ticket, error)
}

type RateRepository interface {
	GetActive(ctx context.Context, vehicleType model.VehicleType) (*model.Rate, error)
	List(ctx context.Context) ([]model.Rate, error)
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
}
