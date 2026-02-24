package workflow

import (
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func OrderCleanupCronWorkflow(ctx workflow.Context) (*entity.CronResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderCleanupCronWorkflow started", "startTime", workflow.Now(ctx))

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		HeartbeatTimeout:    20 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	lastRunTime := time.Time{}
	if workflow.HasLastCompletionResult(ctx) {
		var lastResult entity.CronResult
		if err := workflow.GetLastCompletionResult(ctx, &lastResult); err == nil {
			lastRunTime = lastResult.LastRunTime
			logger.Info("Previous cron run found", "lastRunTime", lastRunTime, "processedCount", lastResult.ProcessedCount)
		}
	}

	_ = lastRunTime 

	var activities *activity.OrderActivities

	var cancelledCount int
	err := workflow.ExecuteActivity(ctx, activities.CleanupStaleOrdersActivity, 360).Get(ctx, &cancelledCount)
	if err != nil {
		logger.Error("CleanupStaleOrdersActivity failed", "error", err)
		return nil, err
	}

	logger.Info("OrderCleanupCronWorkflow completed",
		"cancelledCount", cancelledCount,
		"runTime", workflow.Now(ctx))

	return &entity.CronResult{
		LastRunTime:    workflow.Now(ctx),
		ProcessedCount: cancelledCount,
	}, nil
}
