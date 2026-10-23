package workflow

import (
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	PaidOrderStatusProcessing = "Processing"
	createShipmentChange      = "create-ghn-shipment"
)

func PaidOrderWorkflow(ctx workflow.Context, in entity.PaidOrderInput) (*entity.PaidOrderResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    20,
		},
	})
	var a *activity.PaidOrderActivities

	var record entity.PaidOrderRecord
	if err := workflow.ExecuteActivity(ctx, a.RecordPaidOrder, in).Get(ctx, &record); err != nil {
		return nil, err
	}

	notifyCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	})
	notified := false
	if err := workflow.ExecuteActivity(notifyCtx, a.NotifyOrderPlaced, in, record.Number).Get(ctx, &notified); err != nil {
		workflow.GetLogger(ctx).Warn("order notification failed", "number", record.Number, "error", err)
	}

	if err := workflow.ExecuteActivity(ctx, a.SetPaidOrderStatus, record.Number, PaidOrderStatusProcessing).Get(ctx, nil); err != nil {
		return nil, err
	}
	result := &entity.PaidOrderResult{OrderID: record.ID, Number: record.Number, Status: PaidOrderStatusProcessing, Notified: notified}

	if workflow.GetVersion(ctx, createShipmentChange, workflow.DefaultVersion, 1) == 1 {
		shipCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 45 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    5 * time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    5 * time.Minute,
				MaximumAttempts:    12,
			},
		})
		var shipment entity.ShipmentResult
		if err := workflow.ExecuteActivity(shipCtx, a.CreateShipment, in, record.Number).Get(ctx, &shipment); err != nil {
			workflow.GetLogger(ctx).Warn("shipment request failed", "number", record.Number, "error", err)
			shipment = entity.ShipmentResult{Outcome: entity.ShipmentOutcomeRejected, Reason: err.Error()}
		}
		result.Shipment = &shipment
	}
	return result, nil
}
