package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAdCampaignRepository),
	fx.Provide(NewAdCategoryRepository),
	fx.Provide(NewAdPositionRepository),
	fx.Provide(NewAdRepository),
	fx.Provide(NewAdPositionMappingRepository),
	fx.Provide(NewAdImpressionRepository),
	fx.Provide(NewAdClickRepository),
	fx.Provide(NewAdTargetingRuleRepository),
	fx.Provide(NewAdPerformanceSummaryRepository),
)
