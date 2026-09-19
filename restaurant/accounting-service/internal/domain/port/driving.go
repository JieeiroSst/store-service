package port

import (
	"context"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
)

type AuthCartUsecase interface {
	// PlaceOrder decides accept/reject by operating hours, notifies the
	// downstream services, and records the outcome.
	PlaceOrder(ctx context.Context, cart model.AuthCart) error
}

type PaymentUsecase interface {
	MarkPaid(ctx context.Context, orderID int) error
	Get(ctx context.Context, orderID int) (*model.Payment, error)
}
