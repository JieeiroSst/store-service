package gateway

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/google/uuid"
)

type MockGateway struct{}

func NewMockGateway() port.PaymentGatewayPort {
	return &MockGateway{}
}

func (g *MockGateway) Charge(_ context.Context, req port.ChargeRequest) (port.ChargeResult, error) {
	if req.PaymentMethod == nil {
		return port.ChargeResult{Success: false, ErrorMessage: "no payment method provided"}, nil
	}
	if req.Amount <= 0 {
		return port.ChargeResult{Success: false, ErrorMessage: "invalid charge amount"}, nil
	}

	return port.ChargeResult{
		Success:              true,
		GatewayTransactionID: fmt.Sprintf("mock_%s", uuid.New().String()),
	}, nil
}
