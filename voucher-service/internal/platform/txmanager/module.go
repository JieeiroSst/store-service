package txmanager

import "go.uber.org/fx"

var Module = fx.Module("txmanager",
	fx.Provide(fx.Annotate(NewGormTxManager, fx.As(new(TxManager)))),
)
