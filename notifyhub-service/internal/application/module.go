package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewChannelService),
	fx.Provide(NewDataSourceService),
	fx.Provide(NewTemplateService),
	fx.Provide(NewJobService),
	fx.Provide(NewHistoryService),
	fx.Provide(NewDispatcher),
)
