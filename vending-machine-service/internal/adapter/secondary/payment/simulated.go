package payment

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/google/uuid"
)

type Simulated struct{}

func (Simulated) Charge(ctx context.Context, req port.ChargeRequest) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if req.AmountCents < 0 {
		return "", fmt.Errorf("negative amount %d", req.AmountCents)
	}
	return "sim_" + uuid.NewString(), nil
}

func (Simulated) Refund(ctx context.Context, _ port.RefundRequest) error {
	return ctx.Err()
}

func (Simulated) Lookup(ctx context.Context, _ port.LookupRequest) (string, bool, error) {
	return "", false, ctx.Err()
}
