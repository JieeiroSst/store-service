package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adServingService struct {
	mappingRepo    port.AdPositionMappingRepository
	adRepo         port.AdRepository
	targetingRepo  port.AdTargetingRuleRepository
	impressionRepo port.AdImpressionRepository
	clickRepo      port.AdClickRepository
}

func NewAdServingService(
	mappingRepo port.AdPositionMappingRepository,
	adRepo port.AdRepository,
	targetingRepo port.AdTargetingRuleRepository,
	impressionRepo port.AdImpressionRepository,
	clickRepo port.AdClickRepository,
) port.AdServingUsecase {
	return &adServingService{
		mappingRepo:    mappingRepo,
		adRepo:         adRepo,
		targetingRepo:  targetingRepo,
		impressionRepo: impressionRepo,
		clickRepo:      clickRepo,
	}
}

func (s *adServingService) Serve(ctx context.Context, positionID uint, target model.TargetContext) (*model.ServedAd, error) {
	mappings, err := s.mappingRepo.ListActiveByPosition(ctx, positionID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	var candidates []model.Ad
	var weights []int

	for _, mapping := range mappings {
		ad, err := s.adRepo.GetByID(ctx, mapping.AdID)
		if err != nil || !ad.IsActive || !ad.IsWithinSchedule(now) {
			continue
		}

		rules, err := s.targetingRepo.ListActiveByAd(ctx, ad.ID)
		if err != nil || !matchesTargeting(rules, target) {
			continue
		}

		weight := mapping.Weight
		if weight <= 0 {
			weight = 1
		}
		candidates = append(candidates, *ad)
		weights = append(weights, weight)
	}

	if len(candidates) == 0 {
		return nil, common.ErrNoAdAvailable
	}

	selected := pickWeighted(candidates, weights)

	impression := &model.AdImpression{
		AdID:        selected.ID,
		UserID:      target.UserID,
		SessionID:   target.SessionID,
		IPAddress:   target.IPAddress,
		UserAgent:   target.UserAgent,
		ReferrerURL: target.ReferrerURL,
		PageURL:     target.PageURL,
		PositionID:  &positionID,
	}
	if err := s.impressionRepo.Create(ctx, impression); err != nil {
		return nil, err
	}

	return &model.ServedAd{Ad: selected, ImpressionID: impression.ID}, nil
}

func (s *adServingService) TrackClick(ctx context.Context, adID uint, impressionID uint, target model.TargetContext) (string, error) {
	ad, err := s.adRepo.GetByID(ctx, adID)
	if err != nil {
		return "", err
	}

	var impID *uint
	if impressionID != 0 {
		impID = &impressionID
	}

	click := &model.AdClick{
		AdID:         adID,
		ImpressionID: impID,
		UserID:       target.UserID,
		SessionID:    target.SessionID,
		IPAddress:    target.IPAddress,
		UserAgent:    target.UserAgent,
		ReferrerURL:  target.ReferrerURL,
		TargetURL:    ad.TargetURL,
	}
	if err := s.clickRepo.Create(ctx, click); err != nil {
		return "", err
	}

	return ad.TargetURL, nil
}
