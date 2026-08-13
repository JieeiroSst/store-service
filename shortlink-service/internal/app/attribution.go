package app

import (
	"context"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
)

type AttributionService struct {
	fingerprints ports.FingerprintRepository
	installs     ports.InstallEventRepository
	links        ports.LinkRepository
	webhooks     ports.WebhookRepository
	trigger      *WebhookTrigger
}

func NewAttributionService(
	fingerprints ports.FingerprintRepository,
	installs ports.InstallEventRepository,
	links ports.LinkRepository,
	webhooks ports.WebhookRepository,
	trigger *WebhookTrigger,
) *AttributionService {
	return &AttributionService{fingerprints, installs, links, webhooks, trigger}
}

const maxAttributionWindowHours = 2160

// MatchInstallToClick mirrors matchInstallToClick().
func (s *AttributionService) MatchInstallToClick(ctx context.Context, install domain.FingerprintData, attributionWindowHours int) (*domain.FingerprintMatch, error) {
	if attributionWindowHours <= 0 {
		attributionWindowHours = domain.DefaultAttributionWindowHours
	}

	cutoff := time.Now().Add(-time.Duration(maxAttributionWindowHours) * time.Hour)
	candidates, err := s.fingerprints.CandidateClicks(ctx, cutoff)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	installTime := time.Now()
	var best *domain.FingerprintMatch
	highest := 0

	for _, c := range candidates {
		linkWindow := c.AttributionWindowHours
		if linkWindow <= 0 {
			linkWindow = domain.DefaultAttributionWindowHours
		}
		diffHours := installTime.Sub(c.ClickedAt).Hours()
		if diffHours > float64(linkWindow) {
			continue
		}

		score, matchedFactors := domain.CalculateConfidenceScore(install, c.Fingerprint)
		if score > highest && score >= domain.ConfidenceThreshold {
			highest = score
			best = &domain.FingerprintMatch{
				ClickID:         c.ClickID,
				LinkID:          c.LinkID,
				ConfidenceScore: score,
				MatchedFactors:  matchedFactors,
				ClickedAt:       c.ClickedAt,
			}
		}
	}

	return best, nil
}

// RecordInstallEventResult mirrors recordInstallEvent()'s return shape.
type RecordInstallEventResult struct {
	InstallID    string
	Match        *domain.FingerprintMatch
	DeepLinkData map[string]interface{}
}

func (s *AttributionService) RecordInstallEvent(
	ctx context.Context,
	fp domain.FingerprintData,
	deviceID string,
	attributionWindowHours int,
	sdkName, sdkVersion string,
) (*RecordInstallEventResult, error) {
	if attributionWindowHours <= 0 {
		attributionWindowHours = domain.DefaultAttributionWindowHours
	}

	fingerprintHash := domain.GenerateFingerprintHash(fp)

	match, err := s.MatchInstallToClick(ctx, fp, attributionWindowHours)
	if err != nil {
		return nil, err
	}

	attributionMethod := "none"
	var matchedFactors []string
	var linkID, clickID *string
	var confidenceScore *float64
	if match != nil {
		attributionMethod = "fingerprint"
		matchedFactors = match.MatchedFactors
		linkID = &match.LinkID
		clickID = &match.ClickID
		score := float64(match.ConfidenceScore)
		confidenceScore = &score
	}

	event := &domain.InstallEvent{
		LinkID:                 linkID,
		ClickID:                clickID,
		FingerprintHash:        fingerprintHash,
		ConfidenceScore:        confidenceScore,
		AttributionMethod:      &attributionMethod,
		MatchedFactors:         matchedFactors,
		DeepLinkRetrieved:      false,
		DeepLinkData:           map[string]interface{}{},
		AttributionWindowHours: attributionWindowHours,
		IPAddress:              strPtr(fp.IPAddress),
		UserAgent:              strPtr(fp.UserAgent),
		Timezone:               strPtr(fp.Timezone),
		Language:               strPtr(fp.Language),
		ScreenWidth:            fp.ScreenWidth,
		ScreenHeight:           fp.ScreenHeight,
		Platform:               strPtr(fp.Platform),
		PlatformVersion:        strPtr(fp.PlatformVersion),
		DeviceID:               strPtr(deviceID),
		SDKName:                strPtr(sdkName),
		SDKVersion:             strPtr(sdkVersion),
	}

	installID, err := s.installs.Insert(ctx, event)
	if err != nil {
		return nil, err
	}

	deepLinkData := map[string]interface{}{}

	if match != nil {
		link, err := s.links.GetByID(ctx, match.LinkID, ports.LinkFilter{})
		if err == nil && link != nil {
			deepLinkData = map[string]interface{}{
				"shortCode":          link.ShortCode,
				"originalUrl":        link.OriginalURL,
				"iosUrl":             derefOrEmpty(link.IOSAppStoreURL),
				"androidUrl":         derefOrEmpty(link.AndroidAppStoreURL),
				"webFallbackUrl":     derefOrEmpty(link.WebFallbackURL),
				"utmParameters":      link.UTMParameters,
				"targetingRules":     link.TargetingRules,
				"deepLinkParameters": link.DeepLinkParameters,
				"clickedAt":          match.ClickedAt,
				"confidenceScore":    match.ConfidenceScore,
				"matchedFactors":     match.MatchedFactors,
			}

			if err := s.installs.SetDeepLinkData(ctx, installID, deepLinkData); err != nil {
				return nil, err
			}

			if link.UserID != nil {
				webhooks, err := s.webhooks.ActiveForUser(ctx, *link.UserID)
				if err == nil && len(webhooks) > 0 {
					installEventData := map[string]interface{}{
						"id":              installID,
						"linkId":          match.LinkID,
						"fingerprintHash": fingerprintHash,
						"confidenceScore": match.ConfidenceScore,
						"installedAt":     time.Now().UTC().Format(time.RFC3339Nano),
						"deepLinkData":    deepLinkData,
						"ipAddress":       fp.IPAddress,
						"userAgent":       fp.UserAgent,
						"platform":        fp.Platform,
					}
					s.trigger.Trigger(ctx, webhooks, domain.WebhookEventInstall, installID, installEventData)
				}
			}
		}
	}

	return &RecordInstallEventResult{InstallID: installID, Match: match, DeepLinkData: deepLinkData}, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
