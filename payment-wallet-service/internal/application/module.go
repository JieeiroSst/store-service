package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewWalletService),
	fx.Provide(NewTransactionService),
	fx.Provide(NewPaymentMethodService),
	fx.Provide(NewPocketService),
	fx.Provide(NewPaymentRequestService),
)
