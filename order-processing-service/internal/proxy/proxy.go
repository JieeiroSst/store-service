package proxy

import (
	"context"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
)

type PaymentProxy interface {
	ChargePayment(ctx context.Context, req entity.PaymentRequest) (*entity.PaymentResponse, error)
	RefundPayment(ctx context.Context, paymentID string) error
}

type InventoryProxy interface {
	ReserveStock(ctx context.Context, req entity.InventoryReserveRequest) (*entity.InventoryReserveResponse, error)
	ReleaseStock(ctx context.Context, orderID string) error
}

type ShippingProxy interface {
	CreateShipment(ctx context.Context, req entity.ShippingRequest) (*entity.ShippingResponse, error)
	CancelShipment(ctx context.Context, shippingID string) error
}

type NotificationProxy interface {
	Send(ctx context.Context, req entity.NotificationRequest) (*entity.NotificationResponse, error)
}
