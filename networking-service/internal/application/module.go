package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewBlocker),
	fx.Provide(NewAuthorizer),
	fx.Provide(NewCatalogService),
	fx.Provide(NewKVService),
	fx.Provide(NewSessionService),
	fx.Provide(NewIntentionService),
	fx.Provide(NewHealthRunner),
	fx.Provide(NewSnapshotter),
)
