package app

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func OrderWorkflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderWorkflow started", "orderID", orderID)

	ctx1 := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	})

	var validationResult bool
	err := workflow.ExecuteActivity(ctx1, ValidateOrderActivity, orderID).Get(ctx1, &validationResult)
	if err != nil {
		logger.Error("Order validation failed", "error", err)
		return err
	}

	if !validationResult {
		logger.Info("Order validation failed", "orderID", orderID)
		return temporal.NewApplicationError("Order validation failed", "ValidationError")
	}

	var paymentID string
	err = workflow.ExecuteActivity(ctx1, ProcessPaymentActivity, orderID).Get(ctx1, &paymentID)
	if err != nil {
		logger.Error("Payment processing failed", "error", err)
		return err
	}

	var trackingNumber string
	err = workflow.ExecuteActivity(ctx1, FulfillOrderActivity, orderID, paymentID).Get(ctx1, &trackingNumber)
	if err != nil {
		logger.Error("Order fulfillment failed", "error", err)

		var refundSuccess bool
		err2 := workflow.ExecuteActivity(ctx1, RefundPaymentActivity, paymentID).Get(ctx1, &refundSuccess)
		if err2 != nil {
			logger.Error("Payment refund also failed", "error", err2)
		}

		return err
	}

	err = workflow.ExecuteActivity(ctx1, SendConfirmationActivity, orderID, trackingNumber).Get(ctx1, nil)
	if err != nil {
		logger.Error("Failed to send confirmation", "error", err)
	}

	logger.Info("OrderWorkflow completed successfully", "orderID", orderID)
	return nil
}
