package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAddressRepository),
	fx.Provide(NewCustomerRepository),
	fx.Provide(NewPlanRepository),
	fx.Provide(NewSubscriptionRepository),
	fx.Provide(NewInvoiceRepository),
	fx.Provide(NewTransactionRepository),
)
