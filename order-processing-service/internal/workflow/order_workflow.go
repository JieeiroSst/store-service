package workflow

import (
	"fmt"
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/JIeeiroSst/order-processing-service/pkg/constants"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func OrderProcessingWorkflow(ctx workflow.Context, input entity.OrderWorkflowInput) (*entity.OrderWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderProcessingWorkflow started", "orderID", input.Order.ID)

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Duration(constants.ActivityStartToCloseTimeout) * time.Second,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var currentStatus string = constants.OrderStatusPending
	err := workflow.SetQueryHandler(ctx, "order_status", func() (string, error) {
		return currentStatus, nil
	})
	if err != nil {
		logger.Error("Failed to set query handler", "error", err)
	}

	var compensations []func(ctx workflow.Context) error

	// ============================================================
	// Step 1: Validate Order (Activity)
	// ============================================================
	// Note: We use a nil pointer to the activities struct. Temporal SDK resolves
	// the activity method name and dispatches to the registered instance on the Worker.
	// The actual struct with real dependencies lives on the Worker, not the Workflow.
	var activities *activity.OrderActivities

	err = workflow.ExecuteActivity(ctx, activities.ValidateOrderActivity, input.Order).Get(ctx, nil)
	if err != nil {
		logger.Error("Order validation failed", "orderID", input.Order.ID, "error", err)
		return &entity.OrderWorkflowResult{
			OrderID: input.Order.ID,
			Status:  constants.OrderStatusFailed,
			Message: fmt.Sprintf("validation failed: %v", err),
		}, nil
	}
	currentStatus = constants.OrderStatusValidated

	// Update status in DB
	_ = workflow.ExecuteActivity(ctx, activities.UpdateOrderStatusActivity, input.Order.ID, constants.OrderStatusValidated).Get(ctx, nil)

	// ============================================================
	// Step 2: Reserve Inventory (Activity)
	// ============================================================
	inventoryReq := entity.InventoryReserveRequest{
		OrderID: input.Order.ID,
		Items:   input.Order.Items,
	}
	var inventoryResp entity.InventoryReserveResponse
	err = workflow.ExecuteActivity(ctx, activities.ReserveInventoryActivity, inventoryReq).Get(ctx, &inventoryResp)
	if err != nil {
		logger.Error("Inventory reservation failed", "orderID", input.Order.ID, "error", err)
		runCompensations(ctx, compensations)
		return &entity.OrderWorkflowResult{
			OrderID: input.Order.ID,
			Status:  constants.OrderStatusFailed,
			Message: fmt.Sprintf("inventory reservation failed: %v", err),
		}, nil
	}

	// Add inventory release as compensation
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, activities.CancelOrderActivity, input.Order.ID, "inventory rollback").Get(ctx, nil)
	})

	// ============================================================
	// Step 3: Process Payment (Child Workflow)
	// ============================================================
	paymentReq := entity.PaymentRequest{
		OrderID:    input.Order.ID,
		CustomerID: input.Order.CustomerID,
		Amount:     input.Order.TotalAmount,
		Currency:   input.Order.Currency,
	}

	childPaymentOpts := workflow.ChildWorkflowOptions{
		WorkflowID:         fmt.Sprintf("payment-child-%s", input.Order.ID),
		WorkflowRunTimeout: time.Duration(constants.ChildWorkflowTimeout) * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	childPaymentCtx := workflow.WithChildOptions(ctx, childPaymentOpts)

	var paymentResp entity.PaymentResponse
	err = workflow.ExecuteChildWorkflow(childPaymentCtx, PaymentChildWorkflow, paymentReq).Get(childPaymentCtx, &paymentResp)
	if err != nil {
		logger.Error("Payment child workflow failed", "orderID", input.Order.ID, "error", err)
		runCompensations(ctx, compensations)
		return &entity.OrderWorkflowResult{
			OrderID: input.Order.ID,
			Status:  constants.OrderStatusFailed,
			Message: fmt.Sprintf("payment failed: %v", err),
		}, nil
	}
	currentStatus = constants.OrderStatusPaymentOK

	_ = workflow.ExecuteActivity(ctx, activities.UpdateOrderStatusActivity, input.Order.ID, constants.OrderStatusPaymentOK).Get(ctx, nil)

	// Add payment refund as compensation
	compensations = append(compensations, func(ctx workflow.Context) error {
		return workflow.ExecuteActivity(ctx, activities.CancelOrderActivity, input.Order.ID, "payment refund").Get(ctx, nil)
	})

	// ============================================================
	// Step 4: Create Shipment (Child Workflow)
	// ============================================================
	shippingReq := entity.ShippingRequest{
		OrderID:    input.Order.ID,
		CustomerID: input.Order.CustomerID,
		Address:    input.Address,
	}

	childShippingOpts := workflow.ChildWorkflowOptions{
		WorkflowID:         fmt.Sprintf("shipping-child-%s", input.Order.ID),
		WorkflowRunTimeout: time.Duration(constants.ChildWorkflowTimeout) * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	childShippingCtx := workflow.WithChildOptions(ctx, childShippingOpts)

	var shippingResp entity.ShippingResponse
	err = workflow.ExecuteChildWorkflow(childShippingCtx, ShippingChildWorkflow, shippingReq).Get(childShippingCtx, &shippingResp)
	if err != nil {
		logger.Error("Shipping child workflow failed", "orderID", input.Order.ID, "error", err)
		runCompensations(ctx, compensations)
		return &entity.OrderWorkflowResult{
			OrderID: input.Order.ID,
			Status:  constants.OrderStatusFailed,
			Message: fmt.Sprintf("shipping failed: %v", err),
		}, nil
	}
	currentStatus = constants.OrderStatusShipping

	_ = workflow.ExecuteActivity(ctx, activities.UpdateOrderStatusActivity, input.Order.ID, constants.OrderStatusShipping).Get(ctx, nil)

	// ============================================================
	// Step 5: Send Notification (Activity - non-critical)
	// ============================================================
	notifyReq := entity.NotificationRequest{
		CustomerID: input.Order.CustomerID,
		OrderID:    input.Order.ID,
		Type:       "email",
		Message:    fmt.Sprintf("Your order %s has been shipped! Tracking: %s", input.Order.ID, shippingResp.TrackingCode),
	}
	var notifyResp entity.NotificationResponse
	_ = workflow.ExecuteActivity(ctx, activities.SendNotificationActivity, notifyReq).Get(ctx, &notifyResp)

	// ============================================================
	// Step 6: Mark Order as Completed
	// ============================================================
	currentStatus = constants.OrderStatusCompleted
	_ = workflow.ExecuteActivity(ctx, activities.UpdateOrderStatusActivity, input.Order.ID, constants.OrderStatusCompleted).Get(ctx, nil)

	logger.Info("OrderProcessingWorkflow completed successfully", "orderID", input.Order.ID)

	return &entity.OrderWorkflowResult{
		OrderID:      input.Order.ID,
		Status:       constants.OrderStatusCompleted,
		PaymentID:    paymentResp.PaymentID,
		ShippingID:   shippingResp.ShippingID,
		TrackingCode: shippingResp.TrackingCode,
		Message:      "Order processed successfully",
	}, nil
}

func runCompensations(ctx workflow.Context, compensations []func(ctx workflow.Context) error) {
	logger := workflow.GetLogger(ctx)
	for i := len(compensations) - 1; i >= 0; i-- {
		if err := compensations[i](ctx); err != nil {
			logger.Error("Compensation failed", "index", i, "error", err)
		}
	}
}
