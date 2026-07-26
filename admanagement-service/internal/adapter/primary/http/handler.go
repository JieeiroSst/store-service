package http

import "github.com/JIeeiroSst/admanagement-service/internal/domain/port"

type Handler struct {
	campaign           port.AdCampaignUsecase
	category           port.AdCategoryUsecase
	position           port.AdPositionUsecase
	ad                 port.AdUsecase
	positionMapping    port.AdPositionMappingUsecase
	impression         port.AdImpressionUsecase
	click              port.AdClickUsecase
	targetingRule      port.AdTargetingRuleUsecase
	performanceSummary port.AdPerformanceSummaryUsecase
	serving            port.AdServingUsecase
}

func NewHandler(
	campaign port.AdCampaignUsecase,
	category port.AdCategoryUsecase,
	position port.AdPositionUsecase,
	ad port.AdUsecase,
	positionMapping port.AdPositionMappingUsecase,
	impression port.AdImpressionUsecase,
	click port.AdClickUsecase,
	targetingRule port.AdTargetingRuleUsecase,
	performanceSummary port.AdPerformanceSummaryUsecase,
	serving port.AdServingUsecase,
) *Handler {
	return &Handler{
		campaign:           campaign,
		category:           category,
		position:           position,
		ad:                 ad,
		positionMapping:    positionMapping,
		impression:         impression,
		click:              click,
		targetingRule:      targetingRule,
		performanceSummary: performanceSummary,
		serving:            serving,
	}
}
