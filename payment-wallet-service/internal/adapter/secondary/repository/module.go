package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewWalletRepository),
	fx.Provide(NewTransactionRepository),
	fx.Provide(NewTransferRepository),
	fx.Provide(NewPaymentMethodRepository),
	fx.Provide(NewPocketRepository),
	fx.Provide(NewPaymentRequestRepository),
)
