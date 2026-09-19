package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
	"github.com/JIeeiroSst/parking-lot-service/internal/domain/port"
	"github.com/google/uuid"
)

type parkingService struct {
	vehicles port.VehicleRepository
	spots    port.ParkingSpotRepository
	tickets  port.TicketRepository
	rates    port.RateRepository
	payments port.PaymentRepository
}

func NewParkingService(
	vehicles port.VehicleRepository,
	spots port.ParkingSpotRepository,
	tickets port.TicketRepository,
	rates port.RateRepository,
	payments port.PaymentRepository,
) port.ParkingUsecase {
	return &parkingService{vehicles: vehicles, spots: spots, tickets: tickets, rates: rates, payments: payments}
}

func (s *parkingService) CheckIn(ctx context.Context, plate string, vehicleType model.VehicleType, entryGateID *string) (*model.Ticket, error) {
	if !vehicleType.Valid() {
		return nil, port.ErrInvalidVehicleType
	}

	if _, err := s.vehicles.GetOrCreate(ctx, plate, vehicleType); err != nil {
		return nil, fmt.Errorf("get or create vehicle: %w", err)
	}

	spot, err := s.spots.FindAndReserveAvailable(ctx, model.CompatibleSpotTypes(vehicleType))
	if err != nil {
		return nil, err
	}

	ticket := &model.Ticket{
		TicketID:     uuid.NewString(),
		VehiclePlate: plate,
		SpotID:       spot.SpotID,
		EntryGateID:  entryGateID,
		ParkedTime:   time.Now(),
		Status:       model.TicketActive,
	}
	if err := s.tickets.Create(ctx, ticket); err != nil {
		// Best-effort: don't strand the spot as permanently occupied just
		// because the ticket insert failed after it was claimed.
		_ = s.spots.Release(ctx, spot.SpotID)
		return nil, fmt.Errorf("create ticket: %w", err)
	}
	return ticket, nil
}

func (s *parkingService) CheckOut(ctx context.Context, ticketID string, method model.PaymentMethod, exitGateID *string) (*model.Ticket, *model.Payment, error) {
	if !method.Valid() {
		return nil, nil, port.ErrInvalidPaymentMethod
	}

	ticket, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		return nil, nil, fmt.Errorf("get ticket: %w", err)
	}
	if ticket.Status == model.TicketCompleted {
		return nil, nil, port.ErrTicketAlreadyClosed
	}

	vehicle, err := s.vehicles.Get(ctx, ticket.VehiclePlate)
	if err != nil {
		return nil, nil, fmt.Errorf("get vehicle: %w", err)
	}

	rate, err := s.rates.GetActive(ctx, vehicle.Type)
	if err != nil {
		return nil, nil, fmt.Errorf("get active rate: %w", err)
	}

	now := time.Now()
	ticket.LeaveTime = &now
	ticket.ExitGateID = exitGateID
	ticket.Status = model.TicketCompleted
	if err := s.tickets.Update(ctx, ticket); err != nil {
		return nil, nil, fmt.Errorf("update ticket: %w", err)
	}

	if err := s.spots.Release(ctx, ticket.SpotID); err != nil {
		return nil, nil, fmt.Errorf("release spot: %w", err)
	}

	payment := &model.Payment{
		PaymentID: uuid.NewString(),
		TicketID:  ticket.TicketID,
		Amount:    rate.Calculate(ticket.Duration()),
		Method:    method,
		PaidAt:    &now,
		Status:    model.PaymentPaid,
	}
	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, nil, fmt.Errorf("record payment: %w", err)
	}

	return ticket, payment, nil
}

func (s *parkingService) GetTicket(ctx context.Context, ticketID string) (*model.Ticket, error) {
	return s.tickets.GetByID(ctx, ticketID)
}

func (s *parkingService) ListHistory(ctx context.Context, plate string, limit, offset int) ([]model.Ticket, error) {
	return s.tickets.ListByPlate(ctx, plate, limit, offset)
}

func (s *parkingService) AvailableSpots(ctx context.Context) ([]model.SpotAvailability, error) {
	return s.spots.CountByType(ctx)
}
