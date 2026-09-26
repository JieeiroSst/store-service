package temporalworker

import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/temporalx"
)

func New(c client.Client, queue string, a *Activities) worker.Worker {
	w := worker.New(c, queue, worker.Options{})
	Register(w, a)
	return w
}

func Register(r interface {
	RegisterWorkflowWithOptions(any, workflow.RegisterOptions)
	RegisterActivityWithOptions(any, activity.RegisterOptions)
}, a *Activities) {
	r.RegisterWorkflowWithOptions(OrderWorkflow, workflow.RegisterOptions{Name: temporalx.WorkflowOrder})
	r.RegisterActivityWithOptions(a.AdvanceOrder, activity.RegisterOptions{Name: temporalx.ActivityAdvance})
	r.RegisterActivityWithOptions(a.GenerateOrderDocuments, activity.RegisterOptions{Name: temporalx.ActivityDocuments})
	r.RegisterActivityWithOptions(a.RemindOrder, activity.RegisterOptions{Name: temporalx.ActivityRemind})
}
