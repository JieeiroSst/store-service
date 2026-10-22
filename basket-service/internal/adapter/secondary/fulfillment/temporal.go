package fulfillment

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/JIeeiroSst/basket-service/config"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

const workflowName = "PaidOrderWorkflow"

var ErrDisabled = errors.New("order fulfillment is not configured")

type starter interface {
	ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow any, args ...any) (client.WorkflowRun, error)
}

type Temporal struct {
	client    starter
	taskQueue string
}

func WorkflowID(basketID int) string { return fmt.Sprintf("paid-order-basket-%d", basketID) }

func (t *Temporal) StartPaidOrder(ctx context.Context, order port.PaidOrder) (string, error) {
	if t.client == nil {
		return "", ErrDisabled
	}
	id := WorkflowID(order.BasketID)
	_, err := t.client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                    id,
		TaskQueue:             t.taskQueue,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}, workflowName, order)
	var started *serviceerror.WorkflowExecutionAlreadyStarted
	if err != nil && !errors.As(err, &started) {
		return "", err
	}
	return id, nil
}

func New(lc fx.Lifecycle, cfg *config.Config) (port.Fulfillment, error) {
	t := &Temporal{taskQueue: cfg.Temporal.TaskQueue}
	if t.taskQueue == "" {
		t.taskQueue = "ORDER_PROCESSING_TASK_QUEUE"
	}
	if cfg.Temporal.Address == "" {
		log.Printf("temporal address is not configured: paid baskets will not create orders")
		return t, nil
	}
	c, err := client.NewLazyClient(client.Options{HostPort: cfg.Temporal.Address, Namespace: cfg.Temporal.Namespace})
	if err != nil {
		return nil, fmt.Errorf("temporal client: %w", err)
	}
	t.client = c
	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		c.Close()
		return nil
	}})
	return t, nil
}

var Module = fx.Options(fx.Provide(New))
