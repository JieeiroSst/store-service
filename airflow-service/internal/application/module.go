package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewDAGService),
	fx.Provide(NewDAGRunService),
	fx.Provide(NewTaskInstanceService),
	fx.Provide(NewVariableService),
	fx.Provide(NewPoolService),
	fx.Provide(NewHealthService),
)
