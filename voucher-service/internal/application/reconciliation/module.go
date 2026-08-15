package reconciliation

import "go.uber.org/fx"

var Module = fx.Module("reconciliation-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(ReconciliationService)))),
)
