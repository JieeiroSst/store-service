package app

import "go.uber.org/fx"

var Module = fx.Module("services",
	fx.Provide(
		NewWebhookTrigger,
		NewAttributionService,
		NewClickTracker,
		NewSDKService,
		NewLinkService,
		NewRedirectService,
		NewAnalyticsService,
		NewWebhookService,
		NewTemplateService,
		NewQRService,
		NewHealthService,
	),
)
