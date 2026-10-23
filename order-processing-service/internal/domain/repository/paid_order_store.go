package repository

import (
	"context"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
)

type PaidOrderStore interface {
	CreatePaidOrder(ctx context.Context, in entity.PaidOrderInput) (*entity.PaidOrderRecord, error)
	SetStatus(ctx context.Context, number, status string) error
	SetShipment(ctx context.Context, number string, shipment entity.ShipmentResult) error
}
