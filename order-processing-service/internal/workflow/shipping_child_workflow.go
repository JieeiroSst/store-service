package workflow

import (
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func ShippingChildWorkflow(ctx workflow.Context, req entity.ShippingRequest) (*entity.ShippingResponse, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ShippingChildWorkflow started", "orderID", req.OrderID)

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 45 * time.Second,
		HeartbeatTimeout:    15 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var activities *activity.OrderActivities

	var shippingResp entity.ShippingResponse
	err := workflow.ExecuteActivity(ctx, activities.CreateShipmentActivity, req).Get(ctx, &shippingResp)
	if err != nil {
		logger.Error("ShippingChildWorkflow activity failed", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	notifyReq := entity.NotificationRequest{
		CustomerID: req.CustomerID,
		OrderID:    req.OrderID,
		Type:       "push",
		Message:    "Your order is being prepared for shipment!",
	}
	var notifyResp entity.NotificationResponse
	_ = workflow.ExecuteActivity(ctx, activities.SendNotificationActivity, notifyReq).Get(ctx, &notifyResp)

	logger.Info("ShippingChildWorkflow completed", "orderID", req.OrderID, "shippingID", shippingResp.ShippingID)
	return &shippingResp, nil
}
