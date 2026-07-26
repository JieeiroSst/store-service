package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewSubscriptionService),
	fx.Provide(NewPaymentMethodService),
	fx.Provide(NewInvoiceService),
	fx.Provide(NewTransactionService),
)
