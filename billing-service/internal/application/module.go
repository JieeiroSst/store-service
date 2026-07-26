package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAddressService),
	fx.Provide(NewCustomerService),
	fx.Provide(NewPlanService),
	fx.Provide(NewSubscriptionService),
	fx.Provide(NewInvoiceService),
	fx.Provide(NewTransactionService),
)
