package scheduler

import "go.uber.org/fx"

var Module = fx.Module("scheduler-platform",
	fx.Provide(New),
	fx.Invoke(RegisterLifecycle),
)
