package http

import "github.com/JIeeiroSst/billing-service/internal/domain/port"

type Handler struct {
	address      port.AddressUsecase
	customer     port.CustomerUsecase
	plan         port.PlanUsecase
	subscription port.SubscriptionUsecase
	invoice      port.InvoiceUsecase
	transaction  port.TransactionUsecase
}

func NewHandler(
	address port.AddressUsecase,
	customer port.CustomerUsecase,
	plan port.PlanUsecase,
	subscription port.SubscriptionUsecase,
	invoice port.InvoiceUsecase,
	transaction port.TransactionUsecase,
) *Handler {
	return &Handler{
		address:      address,
		customer:     customer,
		plan:         plan,
		subscription: subscription,
		invoice:      invoice,
		transaction:  transaction,
	}
}
