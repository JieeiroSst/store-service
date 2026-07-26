package http

import (
	"errors"

	"github.com/JIeeiroSst/automatic-payment-service/common"
	gofrhttp "gofr.dev/pkg/gofr/http"

	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
)

type Handler struct {
	subscription  port.SubscriptionUsecase
	paymentMethod port.PaymentMethodUsecase
	invoice       port.InvoiceUsecase
	transaction   port.TransactionUsecase
}

func NewHandler(
	subscription port.SubscriptionUsecase,
	paymentMethod port.PaymentMethodUsecase,
	invoice port.InvoiceUsecase,
	transaction port.TransactionUsecase,
) *Handler {
	return &Handler{
		subscription:  subscription,
		paymentMethod: paymentMethod,
		invoice:       invoice,
		transaction:   transaction,
	}
}

// mapError translates domain errors into gofr errors carrying the right
// HTTP status code; anything unrecognized surfaces as a 500.
func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, common.ErrNotFound):
		return gofrhttp.ErrorEntityNotFound{Name: "id", Value: err.Error()}
	case errors.Is(err, common.ErrInvalidRequest),
		errors.Is(err, common.ErrPaymentMethodRequired):
		return gofrhttp.ErrorInvalidParam{Params: []string{err.Error()}}
	default:
		return err
	}
}
