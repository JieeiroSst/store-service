package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

func newAssignmentTestDeps() (port.DriverAssignmentUsecase, *fakeOrderRepo, *fakeUserClient) {
	orders := newFakeOrderRepo()
	users := newFakeUserClient()
	assignments := newFakeAssignmentRepo()
	svc := NewDriverAssignmentService(assignments, orders, users)
	return svc, orders, users
}

func seedOrder(t *testing.T, orders *fakeOrderRepo) *model.Order {
	t.Helper()
	order, err := orders.Create(context.Background(), &model.Order{CustomerID: "cust-1", RestaurantID: "rest-1", Status: model.OrderStatusReadyForPickup})
	if err != nil {
		t.Fatalf("seed order: %v", err)
	}
	return order
}

func TestAssignDriver(t *testing.T) {
	t.Run("rejects an inactive driver", func(t *testing.T) {
		svc, orders, users := newAssignmentTestDeps()
		order := seedOrder(t, orders)
		users.drivers["drv-1"] = &model.Driver{ID: "drv-1", IsActive: false}

		_, err := svc.AssignDriver(context.Background(), order.ID, "drv-1")
		if !errors.Is(err, port.ErrDriverUnavailable) {
			t.Fatalf("err = %v, want ErrDriverUnavailable", err)
		}
	})

	t.Run("accepting an assignment sets the order's driver", func(t *testing.T) {
		svc, orders, users := newAssignmentTestDeps()
		order := seedOrder(t, orders)
		users.drivers["drv-1"] = &model.Driver{ID: "drv-1", IsActive: true}

		assignment, err := svc.AssignDriver(context.Background(), order.ID, "drv-1")
		if err != nil {
			t.Fatalf("assign driver: %v", err)
		}
		if assignment.Status != model.AssignmentPending {
			t.Fatalf("status = %v, want pending", assignment.Status)
		}

		accepted, err := svc.AcceptAssignment(context.Background(), assignment.ID)
		if err != nil {
			t.Fatalf("accept assignment: %v", err)
		}
		if accepted.Status != model.AssignmentAccepted {
			t.Errorf("status = %v, want accepted", accepted.Status)
		}

		updatedOrder, _ := orders.GetByID(context.Background(), order.ID)
		if updatedOrder.DriverID == nil || *updatedOrder.DriverID != "drv-1" {
			t.Error("expected order.DriverID to be set to the accepting driver")
		}
	})

	t.Run("rejects accepting a non-pending assignment twice", func(t *testing.T) {
		svc, orders, users := newAssignmentTestDeps()
		order := seedOrder(t, orders)
		users.drivers["drv-1"] = &model.Driver{ID: "drv-1", IsActive: true}
		assignment, _ := svc.AssignDriver(context.Background(), order.ID, "drv-1")

		if _, err := svc.AcceptAssignment(context.Background(), assignment.ID); err != nil {
			t.Fatalf("first accept: %v", err)
		}
		if _, err := svc.AcceptAssignment(context.Background(), assignment.ID); !errors.Is(err, port.ErrAssignmentNotPending) {
			t.Fatalf("err = %v, want ErrAssignmentNotPending", err)
		}
	})

	t.Run("reject records a reason and leaves the order unassigned", func(t *testing.T) {
		svc, orders, users := newAssignmentTestDeps()
		order := seedOrder(t, orders)
		users.drivers["drv-1"] = &model.Driver{ID: "drv-1", IsActive: true}
		assignment, _ := svc.AssignDriver(context.Background(), order.ID, "drv-1")

		rejected, err := svc.RejectAssignment(context.Background(), assignment.ID, "too far")
		if err != nil {
			t.Fatalf("reject assignment: %v", err)
		}
		if rejected.Status != model.AssignmentRejected || rejected.RejectionReason != "too far" {
			t.Errorf("unexpected rejected assignment: %+v", rejected)
		}

		updatedOrder, _ := orders.GetByID(context.Background(), order.ID)
		if updatedOrder.DriverID != nil {
			t.Error("expected order.DriverID to stay unset after a rejection")
		}
	})
}
