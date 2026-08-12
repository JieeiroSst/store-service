package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewProjectRepository,
		NewEnvironmentRepository,
		NewFeatureFlagRepository,
		NewFeatureFlagEnvironmentRepository,
		NewStrategyRepository,
		NewConstraintRepository,
		NewUserRepository,
		NewRoleRepository,
		NewMembershipRepository,
		NewTokenRepository,
		NewAuditRepository,
		NewMetricsRepository,
	),
)
