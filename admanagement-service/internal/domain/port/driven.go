package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

type AdCampaignRepository interface {
	Create(ctx context.Context, campaign *model.AdCampaign) error
	GetByID(ctx context.Context, id uint) (*model.AdCampaign, error)
	Update(ctx context.Context, campaign *model.AdCampaign) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdCampaign, error)
}

type AdCategoryRepository interface {
	Create(ctx context.Context, category *model.AdCategory) error
	GetByID(ctx context.Context, id uint) (*model.AdCategory, error)
	Update(ctx context.Context, category *model.AdCategory) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdCategory, error)
}

type AdPositionRepository interface {
	Create(ctx context.Context, position *model.AdPosition) error
	GetByID(ctx context.Context, id uint) (*model.AdPosition, error)
	Update(ctx context.Context, position *model.AdPosition) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdPosition, error)
}

type AdRepository interface {
	Create(ctx context.Context, ad *model.Ad) error
	GetByID(ctx context.Context, id uint) (*model.Ad, error)
	Update(ctx context.Context, ad *model.Ad) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.Ad, error)
	ListByCampaign(ctx context.Context, campaignID uint) ([]model.Ad, error)
}

type AdPositionMappingRepository interface {
	Create(ctx context.Context, mapping *model.AdPositionMapping) error
	GetByID(ctx context.Context, id uint) (*model.AdPositionMapping, error)
	Update(ctx context.Context, mapping *model.AdPositionMapping) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdPositionMapping, error)
	ListActiveByPosition(ctx context.Context, positionID uint) ([]model.AdPositionMapping, error)
}

type AdImpressionRepository interface {
	Create(ctx context.Context, impression *model.AdImpression) error
	GetByID(ctx context.Context, id uint) (*model.AdImpression, error)
	List(ctx context.Context) ([]model.AdImpression, error)
	ListByAd(ctx context.Context, adID uint) ([]model.AdImpression, error)
	CountByAdAndDate(ctx context.Context, adID uint, date time.Time) (int64, error)
}

type AdClickRepository interface {
	Create(ctx context.Context, click *model.AdClick) error
	GetByID(ctx context.Context, id uint) (*model.AdClick, error)
	List(ctx context.Context) ([]model.AdClick, error)
	ListByAd(ctx context.Context, adID uint) ([]model.AdClick, error)
	CountByAdAndDate(ctx context.Context, adID uint, date time.Time) (int64, error)
}

type AdTargetingRuleRepository interface {
	Create(ctx context.Context, rule *model.AdTargetingRule) error
	GetByID(ctx context.Context, id uint) (*model.AdTargetingRule, error)
	Update(ctx context.Context, rule *model.AdTargetingRule) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdTargetingRule, error)
	ListActiveByAd(ctx context.Context, adID uint) ([]model.AdTargetingRule, error)
}

type AdPerformanceSummaryRepository interface {
	Create(ctx context.Context, summary *model.AdPerformanceSummary) error
	GetByID(ctx context.Context, id uint) (*model.AdPerformanceSummary, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdPerformanceSummary, error)
	GetByAdAndDate(ctx context.Context, adID uint, date time.Time) (*model.AdPerformanceSummary, error)
	Upsert(ctx context.Context, summary *model.AdPerformanceSummary) error
	ListByCampaign(ctx context.Context, campaignID uint, from, to time.Time) ([]model.AdPerformanceSummary, error)
}
