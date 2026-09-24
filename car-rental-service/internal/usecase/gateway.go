package usecase

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/model"
)

type ChargeRequest struct {
	Method        model.PaymentMethod
	Amount        float64
	TransactionID string
}

type PaymentGateway interface {
	Charge(ctx context.Context, req ChargeRequest) (model.PaymentStatus, error)
}

type ManualGateway struct{}

func (ManualGateway) Charge(context.Context, ChargeRequest) (model.PaymentStatus, error) {
	return model.PaymentStatusCompleted, nil
}
