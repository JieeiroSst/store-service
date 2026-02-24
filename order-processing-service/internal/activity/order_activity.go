package activity

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/usecase"
	"go.temporal.io/sdk/activity"
)

type OrderActivities struct {
	OrderUseCase usecase.OrderUseCase
}

func NewOrderActivities(uc usecase.OrderUseCase) *OrderActivities {
	return &OrderActivities{OrderUseCase: uc}
}

func (a *OrderActivities) ValidateOrderActivity(ctx context.Context, order entity.Order) error {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidateOrderActivity started", "orderID", order.ID)

	if err := a.OrderUseCase.ValidateOrder(ctx, order); err != nil {
		logger.Error("Order validation failed", "orderID", order.ID, "error", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	logger.Info("ValidateOrderActivity completed", "orderID", order.ID)
	return nil
}

func (a *OrderActivities) ReserveInventoryActivity(ctx context.Context, req entity.InventoryReserveRequest) (*entity.InventoryReserveResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ReserveInventoryActivity started", "orderID", req.OrderID, "itemCount", len(req.Items))

	activity.RecordHeartbeat(ctx, "reserving inventory")

	resp, err := a.OrderUseCase.ReserveInventory(ctx, req)
	if err != nil {
		logger.Error("Inventory reservation failed", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	logger.Info("ReserveInventoryActivity completed", "orderID", req.OrderID, "reserved", resp.Reserved)
	return resp, nil
}

func (a *OrderActivities) ProcessPaymentActivity(ctx context.Context, req entity.PaymentRequest) (*entity.PaymentResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ProcessPaymentActivity started", "orderID", req.OrderID, "amount", req.Amount)

	activity.RecordHeartbeat(ctx, "processing payment")

	resp, err := a.OrderUseCase.ProcessPayment(ctx, req)
	if err != nil {
		logger.Error("Payment processing failed", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	logger.Info("ProcessPaymentActivity completed", "orderID", req.OrderID, "paymentID", resp.PaymentID)
	return resp, nil
}

func (a *OrderActivities) CreateShipmentActivity(ctx context.Context, req entity.ShippingRequest) (*entity.ShippingResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateShipmentActivity started", "orderID", req.OrderID)

	activity.RecordHeartbeat(ctx, "creating shipment")

	resp, err := a.OrderUseCase.CreateShipment(ctx, req)
	if err != nil {
		logger.Error("Shipment creation failed", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	logger.Info("CreateShipmentActivity completed", "orderID", req.OrderID, "shippingID", resp.ShippingID)
	return resp, nil
}

func (a *OrderActivities) SendNotificationActivity(ctx context.Context, req entity.NotificationRequest) (*entity.NotificationResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SendNotificationActivity started", "orderID", req.OrderID, "type", req.Type)

	resp, err := a.OrderUseCase.SendNotification(ctx, req)
	if err != nil {
		logger.Warn("Notification send failed (non-critical)", "orderID", req.OrderID, "error", err)
		return &entity.NotificationResponse{Sent: false}, nil
	}

	logger.Info("SendNotificationActivity completed", "orderID", req.OrderID, "sent", resp.Sent)
	return resp, nil
}

func (a *OrderActivities) UpdateOrderStatusActivity(ctx context.Context, orderID, status string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("UpdateOrderStatusActivity started", "orderID", orderID, "status", status)

	if err := a.OrderUseCase.UpdateOrderStatus(ctx, orderID, status); err != nil {
		logger.Error("Update order status failed", "orderID", orderID, "error", err)
		return err
	}

	logger.Info("UpdateOrderStatusActivity completed", "orderID", orderID, "status", status)
	return nil
}

func (a *OrderActivities) CancelOrderActivity(ctx context.Context, orderID, reason string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("CancelOrderActivity started", "orderID", orderID, "reason", reason)

	if err := a.OrderUseCase.CancelOrder(ctx, orderID, reason); err != nil {
		logger.Error("Cancel order failed", "orderID", orderID, "error", err)
		return err
	}

	logger.Info("CancelOrderActivity completed", "orderID", orderID)
	return nil
}

func (a *OrderActivities) CleanupStaleOrdersActivity(ctx context.Context, olderThanMinutes int) (int, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CleanupStaleOrdersActivity started", "olderThanMinutes", olderThanMinutes)

	activity.RecordHeartbeat(ctx, "cleaning up stale orders")

	count, err := a.OrderUseCase.CleanupStaleOrders(ctx, olderThanMinutes)
	if err != nil {
		logger.Error("Cleanup stale orders failed", "error", err)
		return 0, err
	}

	logger.Info("CleanupStaleOrdersActivity completed", "cancelledCount", count)
	return count, nil
}
