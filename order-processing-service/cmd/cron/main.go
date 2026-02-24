package main

import (
	"context"
	"log"
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/config"
	infraTemporal "github.com/JIeeiroSst/order-processing-service/internal/infrastructure/temporal"
	wf "github.com/JIeeiroSst/order-processing-service/internal/workflow"
	"github.com/JIeeiroSst/order-processing-service/pkg/constants"
	"github.com/JIeeiroSst/order-processing-service/pkg/logger"
	"go.temporal.io/sdk/client"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	defer logger.Sync()

	// Create Temporal client
	temporalClient, err := infraTemporal.NewClient(cfg.Temporal)
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	// Start the Cron Workflow.
	//
	// Temporal Cron Job:
	//   - CronSchedule specifies when each Run is spawned (e.g., "0 */6 * * *" = every 6 hours)
	//   - The Server spawns the first Run immediately, but applies firstWorkflowTaskBackoff
	//     so the first Workflow Task doesn't execute until the scheduled time
	//   - After each Run completes/fails/times out, the next Run is spawned per the schedule
	//   - The Cron Job continues until the Workflow is Terminated or WorkflowExecutionTimeout is reached
	//   - Data can be passed between Runs using HasLastCompletionResult/GetLastCompletionResult
	//
	// Note: Temporal recommends using Schedules over Cron Jobs for new implementations,
	// but Cron Jobs are still fully supported.
	workflowOptions := client.StartWorkflowOptions{
		ID:                       constants.CronWorkflowID,
		TaskQueue:                constants.OrderProcessingTaskQueue,
		CronSchedule:            constants.OrderCleanupCronSchedule,
		WorkflowExecutionTimeout: 365 * 24 * time.Hour, // Run for up to 1 year
		WorkflowRunTimeout:       5 * time.Minute,       // Each cron run has 5 min max
	}

	workflowRun, err := temporalClient.ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		wf.OrderCleanupCronWorkflow,
	)
	if err != nil {
		log.Fatalf("Failed to start cron workflow: %v", err)
	}

	logger.Logger.Infow("Cron Workflow started",
		"workflowID", workflowRun.GetID(),
		"runID", workflowRun.GetRunID(),
		"schedule", constants.OrderCleanupCronSchedule,
	)

	// The cron workflow runs indefinitely. We don't wait for it.
	// To stop it, use: temporal workflow terminate --workflow-id order-cleanup-cron
	logger.Logger.Info("Cron workflow is running. Use 'temporal workflow terminate' to stop it.")
}
