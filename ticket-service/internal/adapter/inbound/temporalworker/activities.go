package temporalworker

import (
	"context"
	"errors"

	"go.temporal.io/sdk/temporal"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/temporalx"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const errPermanent = "permanent"

type Activities struct {
	Orders    inbound.OrderUseCase
	Documents inbound.DocumentUseCase
}

func classify(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrInvalid), errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrForbidden):
		return temporal.NewNonRetryableApplicationError(err.Error(), errPermanent, err)
	}
	return err
}

func (a *Activities) AdvanceOrder(ctx context.Context, orderID int64) (temporalx.Progress, error) {
	p, err := a.Orders.Advance(ctx, orderID)
	if err != nil {
		return temporalx.Progress{}, classify(err)
	}
	return temporalx.Progress{Status: p.Status, ExpiresAt: p.ExpiresAt, RetryAfter: p.RetryAfter, EventStartsAt: p.EventStartsAt, EventCancelled: p.EventCancelled}, nil
}

func (a *Activities) GenerateOrderDocuments(ctx context.Context, orderID int64) error {
	_, err := a.Documents.Generate(ctx, orderID)
	return classify(err)
}

func (a *Activities) RemindOrder(ctx context.Context, orderID int64) error {
	_, err := a.Orders.Remind(ctx, orderID)
	return classify(err)
}
