package workflow

import (
	"errors"
	"testing"

	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

func paidInput() entity.PaidOrderInput {
	return entity.PaidOrderInput{
		BasketID: 12, UserID: 7, Currency: "VND", Total: 300000, PaymentReference: "tx-1",
		Lines: []entity.PaidOrderLine{{ProductID: 3, StockrecordID: 9, Quantity: 2, UnitPrice: 150000}},
	}
}

func TestPaidOrderWorkflowRecordsNotifiesAndMarksProcessing(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	var a *activity.PaidOrderActivities
	env.RegisterActivity(a)
	env.OnActivity(a.RecordPaidOrder, mock.Anything, paidInput()).Return(&entity.PaidOrderRecord{ID: 5, Number: "KS00000012", Created: true}, nil).Once()
	env.OnActivity(a.NotifyOrderPlaced, mock.Anything, paidInput(), "KS00000012").Return(true, nil).Once()
	env.OnActivity(a.SetPaidOrderStatus, mock.Anything, "KS00000012", PaidOrderStatusProcessing).Return(nil).Once()

	env.ExecuteWorkflow(PaidOrderWorkflow, paidInput())

	if !env.IsWorkflowCompleted() || env.GetWorkflowError() != nil {
		t.Fatalf("workflow failed: %v", env.GetWorkflowError())
	}
	var res entity.PaidOrderResult
	if err := env.GetWorkflowResult(&res); err != nil {
		t.Fatal(err)
	}
	if res.OrderID != 5 || res.Number != "KS00000012" || res.Status != PaidOrderStatusProcessing || !res.Notified {
		t.Fatalf("unexpected result: %+v", res)
	}
	env.AssertExpectations(t)
}

func TestPaidOrderWorkflowSurvivesNotificationFailure(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	var a *activity.PaidOrderActivities
	env.RegisterActivity(a)
	env.OnActivity(a.RecordPaidOrder, mock.Anything, mock.Anything).Return(&entity.PaidOrderRecord{ID: 5, Number: "KS00000012"}, nil)
	env.OnActivity(a.NotifyOrderPlaced, mock.Anything, mock.Anything, mock.Anything).Return(false, errors.New("notification down"))
	env.OnActivity(a.SetPaidOrderStatus, mock.Anything, "KS00000012", PaidOrderStatusProcessing).Return(nil).Once()

	env.ExecuteWorkflow(PaidOrderWorkflow, paidInput())

	if env.GetWorkflowError() != nil {
		t.Fatalf("a failed notification must not fail the order: %v", env.GetWorkflowError())
	}
	var res entity.PaidOrderResult
	_ = env.GetWorkflowResult(&res)
	if res.Notified {
		t.Fatal("notified must be false")
	}
}

func TestPaidOrderWorkflowStopsOnInvalidInput(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(&activity.PaidOrderActivities{})

	in := paidInput()
	in.PaymentReference = ""
	env.ExecuteWorkflow(PaidOrderWorkflow, in)

	err := env.GetWorkflowError()
	var appErr *temporal.ApplicationError
	if err == nil || !errors.As(err, &appErr) || !appErr.NonRetryable() {
		t.Fatalf("expected a non-retryable validation error, got %v", err)
	}
}
