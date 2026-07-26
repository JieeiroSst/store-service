package airflowclient

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewClient),
	fx.Provide(NewDAGRepository),
	fx.Provide(NewDAGRunRepository),
	fx.Provide(NewTaskInstanceRepository),
	fx.Provide(NewVariableRepository),
	fx.Provide(NewPoolRepository),
	fx.Provide(NewHealthRepository),
)
