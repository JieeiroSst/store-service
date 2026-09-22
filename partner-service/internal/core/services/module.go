package services

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewPartnerService),
	fx.Provide(NewPartnershipService),
	fx.Provide(NewPartnershipsPartnerService),
	fx.Provide(NewProjectService),
)
