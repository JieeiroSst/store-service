package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewNotificationService),
	fx.Provide(NewDevicePolicy),
	fx.Provide(NewUserDeviceService),
	fx.Provide(NewCampaignSettings),
	fx.Provide(NewCampaignService),
	fx.Provide(NewCampaignRunner),
	fx.Provide(NewAuditor),
	fx.Provide(NewAuditService),
)
