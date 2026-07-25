package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAccountRepository),
	fx.Provide(NewTransactionRepository),
)
