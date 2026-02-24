package usecase

import (
	"context"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
)

type OrderUseCase interface {
	ValidateOrder(ctx context.Context, order entity.Order) error
	ProcessPayment(ctx context.Context, req entity.PaymentRequest) (*entity.PaymentResponse, error)
	ReserveInventory(ctx context.Context, req entity.InventoryReserveRequest) (*entity.InventoryReserveResponse, error)
	CreateShipment(ctx context.Context, req entity.ShippingRequest) (*entity.ShippingResponse, error)
	SendNotification(ctx context.Context, req entity.NotificationRequest) (*entity.NotificationResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	CancelOrder(ctx context.Context, orderID, reason string) error
	CleanupStaleOrders(ctx context.Context, olderThanMinutes int) (int, error)
}
