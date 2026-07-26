package application

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adPerformanceSummaryService struct {
	repo           port.AdPerformanceSummaryRepository
	impressionRepo port.AdImpressionRepository
	clickRepo      port.AdClickRepository
}

func NewAdPerformanceSummaryService(
	repo port.AdPerformanceSummaryRepository,
	impressionRepo port.AdImpressionRepository,
	clickRepo port.AdClickRepository,
) port.AdPerformanceSummaryUsecase {
	return &adPerformanceSummaryService{repo: repo, impressionRepo: impressionRepo, clickRepo: clickRepo}
}

func (s *adPerformanceSummaryService) Get(ctx context.Context, id uint) (*model.AdPerformanceSummary, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adPerformanceSummaryService) List(ctx context.Context) ([]model.AdPerformanceSummary, error) {
	return s.repo.List(ctx)
}

// Recompute counts the day's raw ad_impressions/ad_clicks rows and upserts
// the ad_performance_summary rollup, preserving any cost/revenue already
// recorded on it (those come from billing, not from this count).
func (s *adPerformanceSummaryService) Recompute(ctx context.Context, adID uint, date time.Time) (*model.AdPerformanceSummary, error) {
	impressions, err := s.impressionRepo.CountByAdAndDate(ctx, adID, date)
	if err != nil {
		return nil, err
	}
	clicks, err := s.clickRepo.CountByAdAndDate(ctx, adID, date)
	if err != nil {
		return nil, err
	}

	var ctr float64
	if impressions > 0 {
		ctr = float64(clicks) / float64(impressions)
	}

	existing, err := s.repo.GetByAdAndDate(ctx, adID, date)
	if err != nil && !errors.Is(err, common.ErrNotFound) {
		return nil, err
	}

	summary := &model.AdPerformanceSummary{
		AdID:        adID,
		Date:        time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
		Impressions: int(impressions),
		Clicks:      int(clicks),
		CTR:         ctr,
	}
	if existing != nil {
		summary.ID = existing.ID
		summary.Cost = existing.Cost
		summary.Revenue = existing.Revenue
	}

	if err := s.repo.Upsert(ctx, summary); err != nil {
		return nil, err
	}
	return summary, nil
}

// CampaignRollup sums the daily summaries of every ad in a campaign over
// [from, to] and recomputes a blended CTR across the whole range.
func (s *adPerformanceSummaryService) CampaignRollup(ctx context.Context, campaignID uint, from, to time.Time) (*model.CampaignPerformance, error) {
	summaries, err := s.repo.ListByCampaign(ctx, campaignID, from, to)
	if err != nil {
		return nil, err
	}

	rollup := &model.CampaignPerformance{CampaignID: campaignID}
	for _, summary := range summaries {
		rollup.Impressions += summary.Impressions
		rollup.Clicks += summary.Clicks
		rollup.Cost += summary.Cost
		rollup.Revenue += summary.Revenue
	}
	if rollup.Impressions > 0 {
		rollup.CTR = float64(rollup.Clicks) / float64(rollup.Impressions)
	}
	return rollup, nil
}
