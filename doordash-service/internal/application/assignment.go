package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type driverAssignmentService struct {
	assignments port.DriverAssignmentRepository
	orders      port.OrderRepository
	users       port.UserClient
}

func NewDriverAssignmentService(
	assignments port.DriverAssignmentRepository,
	orders port.OrderRepository,
	users port.UserClient,
) port.DriverAssignmentUsecase {
	return &driverAssignmentService{assignments: assignments, orders: orders, users: users}
}

func (s *driverAssignmentService) AssignDriver(ctx context.Context, orderID int64, driverID string) (*model.DriverAssignment, error) {
	if _, err := s.orders.GetByID(ctx, orderID); err != nil {
		return nil, err
	}

	driver, err := s.users.GetDriver(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if !driver.IsActive {
		return nil, port.ErrDriverUnavailable
	}

	return s.assignments.Create(ctx, &model.DriverAssignment{
		DriverID:   driverID,
		OrderID:    orderID,
		Status:     model.AssignmentPending,
		AssignedAt: time.Now(),
	})
}

func (s *driverAssignmentService) AcceptAssignment(ctx context.Context, assignmentID int64) (*model.DriverAssignment, error) {
	assignment, err := s.pendingAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	assignment.Status = model.AssignmentAccepted
	assignment.AcceptedAt = &now
	updated, err := s.assignments.Update(ctx, assignment)
	if err != nil {
		return nil, err
	}

	if err := s.orders.AssignDriver(ctx, assignment.OrderID, assignment.DriverID); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *driverAssignmentService) RejectAssignment(ctx context.Context, assignmentID int64, reason string) (*model.DriverAssignment, error) {
	assignment, err := s.pendingAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	assignment.Status = model.AssignmentRejected
	assignment.RejectionReason = reason
	return s.assignments.Update(ctx, assignment)
}

func (s *driverAssignmentService) CompleteAssignment(ctx context.Context, assignmentID int64) (*model.DriverAssignment, error) {
	assignment, err := s.assignments.GetByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if assignment.Status != model.AssignmentAccepted {
		return nil, port.ErrAssignmentNotPending
	}

	now := time.Now()
	assignment.Status = model.AssignmentCompleted
	assignment.CompletedAt = &now
	return s.assignments.Update(ctx, assignment)
}

func (s *driverAssignmentService) ListAssignmentsByDriver(ctx context.Context, driverID string) ([]model.DriverAssignment, error) {
	return s.assignments.ListByDriver(ctx, driverID)
}

func (s *driverAssignmentService) pendingAssignment(ctx context.Context, assignmentID int64) (*model.DriverAssignment, error) {
	assignment, err := s.assignments.GetByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if assignment.Status != model.AssignmentPending {
		return nil, port.ErrAssignmentNotPending
	}
	return assignment, nil
}
