package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewChannelRepository),
	fx.Provide(NewDataSourceRepository),
	fx.Provide(NewTemplateRepository),
	fx.Provide(NewJobRepository),
	fx.Provide(NewHistoryRepository),
)
