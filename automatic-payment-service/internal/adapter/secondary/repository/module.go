package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewSubscriptionRepository),
	fx.Provide(NewPaymentMethodRepository),
	fx.Provide(NewTransactionRepository),
	fx.Provide(NewInvoiceRepository),
)
