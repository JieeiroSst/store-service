package wallet

import "go.uber.org/fx"

var Module = fx.Module("wallet-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(WalletService)))),
)
