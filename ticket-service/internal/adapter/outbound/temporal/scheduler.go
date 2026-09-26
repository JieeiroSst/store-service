package temporal

import (
	"context"
	"errors"
	"fmt"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"

	"github.com/JIeeiroSst/ticket-service/config"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/temporalx"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
)

type Scheduler struct {
	c     client.Client
	queue string
}

var _ outbound.OrderLifecycle = (*Scheduler)(nil)

func NewScheduler(c client.Client, queue string) *Scheduler { return &Scheduler{c: c, queue: queue} }

func NewClient(cfg config.Config) (client.Client, error) {
	return client.NewLazyClient(client.Options{HostPort: cfg.TemporalAddress, Namespace: cfg.TemporalNamespace})
}

func (s *Scheduler) Track(ctx context.Context, orderID int64, expiresAt time.Time) error {
	_, err := s.c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       temporalx.WorkflowID(s.queue, orderID),
		TaskQueue:                s.queue,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_FAIL,
	}, temporalx.WorkflowOrder, temporalx.OrderInput{OrderID: orderID})
	var started *serviceerror.WorkflowExecutionAlreadyStarted
	if errors.As(err, &started) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("start order workflow: %w", err)
	}
	return nil
}

func (s *Scheduler) Nudge(ctx context.Context, orderID int64) error {
	err := s.c.SignalWorkflow(ctx, temporalx.WorkflowID(s.queue, orderID), "", temporalx.SignalOrderChanged, nil)
	var missing *serviceerror.NotFound
	if errors.As(err, &missing) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("signal order workflow: %w", err)
	}
	return nil
}
