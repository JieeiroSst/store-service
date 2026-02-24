package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/config"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	infraTemporal "github.com/JIeeiroSst/order-processing-service/internal/infrastructure/temporal"
	wf "github.com/JIeeiroSst/order-processing-service/internal/workflow"
	"github.com/JIeeiroSst/order-processing-service/pkg/constants"
	"github.com/JIeeiroSst/order-processing-service/pkg/logger"
	"go.temporal.io/sdk/client"
	enumspb "go.temporal.io/api/enums/v1"
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

	// Build sample order input
	orderInput := entity.OrderWorkflowInput{
		Order: entity.Order{
			ID:         "ORD-20260224-001",
			CustomerID: "CUST-1001",
			Items: []entity.OrderItem{
				{ProductID: "PROD-A1", Name: "Wireless Keyboard", Quantity: 2, Price: 49.99},
				{ProductID: "PROD-B2", Name: "USB-C Hub", Quantity: 1, Price: 29.99},
			},
			TotalAmount: 129.97,
			Currency:    "USD",
			Status:      constants.OrderStatusPending,
			CreatedAt:   time.Now(),
		},
		Address: entity.Address{
			Street:  "123 Main St",
			City:    "Ho Chi Minh City",
			State:   "HCMC",
			ZipCode: "700000",
			Country: "VN",
		},
	}

	// Workflow Execution options
	workflowOptions := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-%s", constants.OrderWorkflowIDPrefix, orderInput.Order.ID),
		TaskQueue:             constants.OrderProcessingTaskQueue,
		WorkflowRunTimeout:    10 * time.Minute,
		WorkflowTaskTimeout:   10 * time.Second,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}

	// Start the Workflow Execution
	// The Temporal Server tracks the progress of the Workflow Execution.
	// The actual code runs on the Worker, not the Server.
	workflowRun, err := temporalClient.ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		wf.OrderProcessingWorkflow,
		orderInput,
	)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	logger.Logger.Infow("Workflow started",
		"workflowID", workflowRun.GetID(),
		"runID", workflowRun.GetRunID(),
	)

	// Wait for workflow completion and get result
	var result entity.OrderWorkflowResult
	err = workflowRun.Get(context.Background(), &result)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	logger.Logger.Infow("Workflow completed",
		"orderID", result.OrderID,
		"status", result.Status,
		"paymentID", result.PaymentID,
		"shippingID", result.ShippingID,
		"trackingCode", result.TrackingCode,
		"message", result.Message,
	)
}
