package scheduler

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Invoke(Register),
)
