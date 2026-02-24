package worker

import (
	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	wf "github.com/JIeeiroSst/order-processing-service/internal/workflow"
	"github.com/JIeeiroSst/order-processing-service/pkg/constants"
	"go.temporal.io/sdk/client"
	sdkworker "go.temporal.io/sdk/worker"
)

func NewOrderWorker(c client.Client, activities *activity.OrderActivities) sdkworker.Worker {
	w := sdkworker.New(c, constants.OrderProcessingTaskQueue, sdkworker.Options{
		MaxConcurrentActivityExecutionSize:     10,
		MaxConcurrentWorkflowTaskExecutionSize: 10,
	})

	w.RegisterWorkflow(wf.OrderProcessingWorkflow)
	w.RegisterWorkflow(wf.PaymentChildWorkflow)
	w.RegisterWorkflow(wf.ShippingChildWorkflow)
	w.RegisterWorkflow(wf.OrderCleanupCronWorkflow)

	w.RegisterActivity(activities)

	return w
}

func NewCronWorker(c client.Client, activities *activity.OrderActivities) sdkworker.Worker {
	w := sdkworker.New(c, constants.CronTaskQueue, sdkworker.Options{
		MaxConcurrentActivityExecutionSize:     5,
		MaxConcurrentWorkflowTaskExecutionSize: 2,
	})

	w.RegisterWorkflow(wf.OrderCleanupCronWorkflow)
	w.RegisterActivity(activities)

	return w
}
