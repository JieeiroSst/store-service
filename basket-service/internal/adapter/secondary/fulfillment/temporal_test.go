package fulfillment

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

type fakeStarter struct {
	options client.StartWorkflowOptions
	name    any
	args    []any
	err     error
}

func (f *fakeStarter) ExecuteWorkflow(_ context.Context, o client.StartWorkflowOptions, wf any, args ...any) (client.WorkflowRun, error) {
	f.options, f.name, f.args = o, wf, args
	return nil, f.err
}

func TestStartPaidOrderUsesAStableWorkflowID(t *testing.T) {
	s := &fakeStarter{}
	tc := &Temporal{client: s, taskQueue: "ORDER_PROCESSING_TASK_QUEUE"}
	order := port.PaidOrder{BasketID: 12, UserID: 7}

	id, err := tc.StartPaidOrder(context.Background(), order)
	if err != nil || id != "paid-order-basket-12" {
		t.Fatalf("start: %q %v", id, err)
	}
	if s.options.ID != id || s.options.TaskQueue != "ORDER_PROCESSING_TASK_QUEUE" || s.options.WorkflowIDReusePolicy != enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE {
		t.Fatalf("unexpected options: %+v", s.options)
	}
	if s.name != "PaidOrderWorkflow" || s.args[0].(port.PaidOrder).BasketID != 12 {
		t.Fatalf("unexpected workflow call: %v %v", s.name, s.args)
	}

	s.err = serviceerror.NewWorkflowExecutionAlreadyStarted("exists", "", "")
	if id, err := tc.StartPaidOrder(context.Background(), order); err != nil || id != "paid-order-basket-12" {
		t.Fatalf("an already started workflow is a success: %q %v", id, err)
	}

	s.err = errors.New("unavailable")
	if _, err := tc.StartPaidOrder(context.Background(), order); err == nil {
		t.Fatal("transport errors must surface")
	}
	if _, err := (&Temporal{}).StartPaidOrder(context.Background(), order); !errors.Is(err, ErrDisabled) {
		t.Fatalf("unconfigured client: %v", err)
	}
}
