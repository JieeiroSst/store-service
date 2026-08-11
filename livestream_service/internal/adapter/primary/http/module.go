package http

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewHandler),
	fx.Provide(NewSRSWebhookHandler),
	fx.Provide(NewInternalHandler),
	fx.Provide(NewWSHandler),
)
