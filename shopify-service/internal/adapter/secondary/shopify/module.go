package shopify

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewClient),
	fx.Provide(NewWebhookVerifier),
)
