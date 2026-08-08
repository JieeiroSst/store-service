package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewBotService),
	fx.Provide(NewPublisherRegistry),
	fx.Provide(NewContentService),
)
