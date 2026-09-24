package orchestrator

import (
	"context"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	tc "github.com/JIeeiroSst/customer-relationship-service/internal/infrastructure/temporal"
	"github.com/JIeeiroSst/customer-relationship-service/internal/workflows"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

const syncTimeout = 3 * time.Second

type temporalOrchestrator struct{ c *tc.Client }

func New(c *tc.Client) port.ContractOrchestrator {
	if !c.Enabled() {
		return noop{}
	}
	return &temporalOrchestrator{c: c}
}

func (o *temporalOrchestrator) Sync(ctx context.Context, contractID uint) error {
	ctx, cancel := context.WithTimeout(ctx, syncTimeout)
	defer cancel()

	_, err := o.c.Client.SignalWithStartWorkflow(ctx,
		workflows.WorkflowID(contractID),
		workflows.SignalStateChanged, nil,
		client.StartWorkflowOptions{
			ID:        workflows.WorkflowID(contractID),
			TaskQueue: o.c.Config.TaskQueue,
		},
		workflows.ContractLifecycle,
		workflows.LifecycleInput{ContractID: contractID, Reminders: o.c.Config.Reminders},
	)
	return err
}

type noop struct{}

func (noop) Sync(context.Context, uint) error { return nil }

var Module = fx.Options(fx.Provide(New))
