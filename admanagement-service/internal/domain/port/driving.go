package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

type AdCampaignUsecase interface {
	Create(ctx context.Context, campaign *model.AdCampaign) error
	Get(ctx context.Context, id uint) (*model.AdCampaign, error)
	Update(ctx context.Context, campaign *model.AdCampaign) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdCampaign, error)
	ChangeStatus(ctx context.Context, id uint, status model.CampaignStatus) (*model.AdCampaign, error)
}

type AdCategoryUsecase interface {
	Create(ctx context.Context, category *model.AdCategory) error
	Get(ctx context.Context, id uint) (*model.AdCategory, error)
	Update(ctx context.Context, category *model.AdCategory) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdCategory, error)
}

type AdPositionUsecase interface {
	Create(ctx context.Context, position *model.AdPosition) error
	Get(ctx context.Context, id uint) (*model.AdPosition, error)
	Update(ctx context.Context, position *model.AdPosition) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdPosition, error)
}

type AdUsecase interface {
	Create(ctx context.Context, ad *model.Ad) error
	Get(ctx context.Context, id uint) (*model.Ad, error)
	Update(ctx context.Context, ad *model.Ad) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.Ad, error)
	ListByCampaign(ctx context.Context, campaignID uint) ([]model.Ad, error)
}

type AdPositionMappingUsecase interface {
	Create(ctx context.Context, mapping *model.AdPositionMapping) error
	Get(ctx context.Context, id uint) (*model.AdPositionMapping, error)
	Update(ctx context.Context, mapping *model.AdPositionMapping) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdPositionMapping, error)
}

type AdImpressionUsecase interface {
	Get(ctx context.Context, id uint) (*model.AdImpression, error)
	List(ctx context.Context) ([]model.AdImpression, error)
	ListByAd(ctx context.Context, adID uint) ([]model.AdImpression, error)
}

type AdClickUsecase interface {
	Get(ctx context.Context, id uint) (*model.AdClick, error)
	List(ctx context.Context) ([]model.AdClick, error)
	ListByAd(ctx context.Context, adID uint) ([]model.AdClick, error)
}

type AdTargetingRuleUsecase interface {
	Create(ctx context.Context, rule *model.AdTargetingRule) error
	Get(ctx context.Context, id uint) (*model.AdTargetingRule, error)
	Update(ctx context.Context, rule *model.AdTargetingRule) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdTargetingRule, error)
}

type AdPerformanceSummaryUsecase interface {
	Get(ctx context.Context, id uint) (*model.AdPerformanceSummary, error)
	List(ctx context.Context) ([]model.AdPerformanceSummary, error)
	Recompute(ctx context.Context, adID uint, date time.Time) (*model.AdPerformanceSummary, error)
	CampaignRollup(ctx context.Context, campaignID uint, from, to time.Time) (*model.CampaignPerformance, error)
}

type AdServingUsecase interface {
	Serve(ctx context.Context, positionID uint, target model.TargetContext) (*model.ServedAd, error)
	TrackClick(ctx context.Context, adID uint, impressionID uint, target model.TargetContext) (targetURL string, err error)
}
