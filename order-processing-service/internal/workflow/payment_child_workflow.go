package workflow

import (
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func PaymentChildWorkflow(ctx workflow.Context, req entity.PaymentRequest) (*entity.PaymentResponse, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("PaymentChildWorkflow started", "orderID", req.OrderID, "amount", req.Amount)

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    20 * time.Second,
			MaximumAttempts:    3,
			NonRetryableErrorTypes: []string{
				"PaymentDeclinedError",
				"InsufficientFundsError",
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var activities *activity.OrderActivities

	var paymentResp entity.PaymentResponse
	err := workflow.ExecuteActivity(ctx, activities.ProcessPaymentActivity, req).Get(ctx, &paymentResp)
	if err != nil {
		logger.Error("PaymentChildWorkflow activity failed", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	signalCh := workflow.GetSignalChannel(ctx, "payment-confirmed")
	signalCh.Receive(ctx, &paymentResp)

	logger.Info("PaymentChildWorkflow completed", "orderID", req.OrderID, "paymentID", paymentResp.PaymentID)
	return &paymentResp, nil
}
