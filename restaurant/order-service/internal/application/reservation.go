package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/order-service/common"
	"github.com/JIeeiroSst/order-service/internal/domain/model"
	"github.com/JIeeiroSst/order-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
)

// reservationConflictWindow: two bookings for the same table within this
// window of each other are treated as a double-booking.
const reservationConflictWindow = 2 * time.Hour

type reservationService struct {
	repo port.ReservationRepository
}

func NewReservationService(repo port.ReservationRepository) port.ReservationUsecase {
	return &reservationService{repo: repo}
}

func (s *reservationService) Book(ctx context.Context, reservation *model.Reservation) error {
	conflicts, err := s.repo.FindConflicting(ctx, reservation.TableName, reservation.ReservedAt, reservationConflictWindow)
	if err != nil {
		return err
	}
	if len(conflicts) > 0 {
		return common.ErrTableAlreadyBooked
	}

	reservation.ID = logger.GearedIntID()
	reservation.Status = model.ReservationStatusBooked
	reservation.CreatedDate = time.Now()

	return s.repo.Create(ctx, *reservation)
}

func (s *reservationService) Cancel(ctx context.Context, id int) error {
	return s.repo.Cancel(ctx, id)
}

func (s *reservationService) FindByID(ctx context.Context, id int) (*model.Reservation, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *reservationService) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	return s.repo.FindAll(ctx, pagination)
}
