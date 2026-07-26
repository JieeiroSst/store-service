package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAdCampaignService),
	fx.Provide(NewAdCategoryService),
	fx.Provide(NewAdPositionService),
	fx.Provide(NewAdService),
	fx.Provide(NewAdPositionMappingService),
	fx.Provide(NewAdImpressionService),
	fx.Provide(NewAdClickService),
	fx.Provide(NewAdTargetingRuleService),
	fx.Provide(NewAdPerformanceSummaryService),
	fx.Provide(NewAdServingService),
)
