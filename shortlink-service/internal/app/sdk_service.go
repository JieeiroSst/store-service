package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/adapters/repo"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
)

var ErrLinkNotFound = errors.New("link not found")
var ErrInstallNotFound = errors.New("install not found")

// SDKService implements the use cases behind /api/sdk/v1/* (src/routes/sdk.ts).
type SDKService struct {
	links        ports.LinkRepository
	installs     ports.InstallEventRepository
	inAppEvents  ports.InAppEventRepository
	clickEvents  ports.ClickEventRepository
	webhooks     ports.WebhookRepository
	attribution  *AttributionService
	clickTracker *ClickTracker
	cache        ports.Cache
	trigger      *WebhookTrigger
}

func NewSDKService(
	links ports.LinkRepository,
	installs ports.InstallEventRepository,
	inAppEvents ports.InAppEventRepository,
	clickEvents ports.ClickEventRepository,
	webhooks ports.WebhookRepository,
	attribution *AttributionService,
	clickTracker *ClickTracker,
	cache ports.Cache,
	trigger *WebhookTrigger,
) *SDKService {
	return &SDKService{links, installs, inAppEvents, clickEvents, webhooks, attribution, clickTracker, cache, trigger}
}

// InstallInput mirrors POST /api/sdk/v1/install's request schema.
type InstallInput struct {
	// TrustedIP is resolved server-side (connection / X-Forwarded-For),
	// NEVER the client-supplied body.ipAddress -- see ResolveClientIP.
	TrustedIP              string
	UserAgent              string
	Timezone               string
	Language               string
	ScreenWidth            *int
	ScreenHeight           *int
	Platform               string
	PlatformVersion        string
	DeviceID               string
	AttributionWindowHours int
	SDKName                string
	SDKVersion             string
}

type InstallResult struct {
	InstallID       string
	Attributed      bool
	ConfidenceScore int
	MatchedFactors  []string
	DeepLinkData    map[string]interface{}
}

// Install mirrors POST /api/sdk/v1/install's handler body.
func (s *SDKService) Install(ctx context.Context, in InstallInput) (*InstallResult, error) {
	fp := domain.FingerprintData{
		IPAddress:       in.TrustedIP,
		UserAgent:       in.UserAgent,
		Timezone:        in.Timezone,
		Language:        in.Language,
		ScreenWidth:     in.ScreenWidth,
		ScreenHeight:    in.ScreenHeight,
		Platform:        in.Platform,
		PlatformVersion: in.PlatformVersion,
	}

	result, err := s.attribution.RecordInstallEvent(ctx, fp, in.DeviceID, in.AttributionWindowHours, in.SDKName, in.SDKVersion)
	if err != nil {
		return nil, err
	}

	confidence := 0
	var matchedFactors []string
	if result.Match != nil {
		confidence = result.Match.ConfidenceScore
		matchedFactors = result.Match.MatchedFactors
	}
	if matchedFactors == nil {
		matchedFactors = []string{}
	}

	return &InstallResult{
		InstallID:       result.InstallID,
		Attributed:      result.Match != nil,
		ConfidenceScore: confidence,
		MatchedFactors:  matchedFactors,
		DeepLinkData:    result.DeepLinkData,
	}, nil
}

// ResolveInput mirrors GET /api/sdk/v1/resolve/:shortCode(/:templateSlug).
type ResolveInput struct {
	ShortCode    string
	TemplateSlug string

	IP             string
	UserAgent      string
	Referrer       string
	AcceptLanguage string
	Method         string

	TrustEdgeBotHeader bool
	EdgeBotHeaderValue string

	UTMSource, UTMMedium, UTMCampaign string
	FPTimezone, FPLanguage            string
	FPScreenWidth, FPScreenHeight     string
}

// ResolveResult mirrors handleResolve()'s JSON response.
type ResolveResult struct {
	ShortCode        string
	LinkID           string
	DeepLinkPath     *string
	AppScheme        *string
	IOSUrl           *string
	AndroidUrl       *string
	WebUrl           *string
	UTMParameters    domain.UTMParameters
	CustomParameters map[string]interface{}
	ClickedAt        time.Time
}

// Resolve mirrors handleResolve() in sdk.ts: cache-or-DB lookup, safety
// gate, async click + fingerprint tracking, and a JSON deep-link response
// (no redirect).
func (s *SDKService) Resolve(ctx context.Context, in ResolveInput) (*ResolveResult, error) {
	link, err := resolveLinkCached(ctx, s.links, s.cache, in.ShortCode, in.TemplateSlug)
	if err != nil {
		return nil, err
	}

	safety := domain.EvaluateLinkSafety(domain.LinkSafetyInput{
		IsActive:         &link.IsActive,
		WarnAt:           link.WarnAt,
		OwnerSuspendedAt: link.OwnerSuspendedAt,
	})
	if safety == domain.SafetyBlock {
		return nil, ErrLinkNotFound
	}

	s.clickTracker.TrackAsync(ClickTrackInput{
		Link: link, IP: in.IP, UserAgent: in.UserAgent, Referrer: in.Referrer,
		AcceptLanguage: in.AcceptLanguage, Method: in.Method,
		TrustEdgeBotHeader: in.TrustEdgeBotHeader, EdgeBotHeaderValue: in.EdgeBotHeaderValue,
		UTMSource: in.UTMSource, UTMMedium: in.UTMMedium, UTMCampaign: in.UTMCampaign,
		FPTimezone: in.FPTimezone, FPLanguage: in.FPLanguage,
		FPScreenWidth: in.FPScreenWidth, FPScreenHeight: in.FPScreenHeight,
		IncludeUTMInWebhookPayload: false,
	})

	customParams := link.DeepLinkParameters
	if customParams == nil {
		customParams = map[string]interface{}{}
	}

	return &ResolveResult{
		ShortCode:        link.ShortCode,
		LinkID:           link.ID,
		DeepLinkPath:     link.DeepLinkPath,
		AppScheme:        link.AppScheme,
		IOSUrl:           link.IOSAppStoreURL,
		AndroidUrl:       link.AndroidAppStoreURL,
		WebUrl:           link.WebFallbackURL,
		UTMParameters:    link.UTMParameters,
		CustomParameters: customParams,
		ClickedAt:        time.Now().UTC(),
	}, nil
}

func resolveLinkCached(ctx context.Context, links ports.LinkRepository, cache ports.Cache, shortCode, templateSlug string) (*domain.Link, error) {
	cacheKey := ports.LinkResolutionCacheKey(shortCode, templateSlug)

	if cache.Enabled() {
		if raw, ok := cache.Get(ctx, cacheKey); ok {
			var cached cachedLink
			if err := json.Unmarshal([]byte(raw), &cached); err == nil {
				return cached.toDomain(), nil
			}
		}
	}

	link, err := links.ResolveForRedirect(ctx, shortCode, templateSlug)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	if cache.Enabled() {
		if raw, err := json.Marshal(newCachedLink(link)); err == nil {
			cache.Set(ctx, cacheKey, string(raw), 5*time.Minute)
		}
	}

	return link, nil
}

// EventInput mirrors POST /api/sdk/v1/event's request schema.
type EventInput struct {
	InstallID         string
	EventName         string
	EventData         map[string]interface{}
	Timestamp         *time.Time
	AttributedLinkID  string
	AttributedClickID string
	LinkOpenedAt      *time.Time
	SessionID         string
	SDKName           string
	SDKVersion        string
}

type EventResult struct {
	EventID      string
	Acknowledged bool
}

// Event mirrors POST /api/sdk/v1/event's handler body.
func (s *SDKService) Event(ctx context.Context, in EventInput) (*EventResult, error) {
	install, err := s.installs.GetByID(ctx, in.InstallID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInstallNotFound
		}
		return nil, err
	}

	ts := time.Now().UTC()
	if in.Timestamp != nil {
		ts = *in.Timestamp
	}
	data := in.EventData
	if data == nil {
		data = map[string]interface{}{}
	}

	event := &domain.InAppEvent{
		InstallID:         in.InstallID,
		EventName:         in.EventName,
		EventData:         data,
		EventTimestamp:    ts,
		AttributedLinkID:  strPtr(in.AttributedLinkID),
		AttributedClickID: strPtr(in.AttributedClickID),
		AttributedAt:      in.LinkOpenedAt,
		SessionID:         strPtr(in.SessionID),
		SDKName:           strPtr(in.SDKName),
		SDKVersion:        strPtr(in.SDKVersion),
	}

	eventID, err := s.inAppEvents.Insert(ctx, event)
	if err != nil {
		return nil, err
	}

	if install.LinkID != nil {
		webhooks, err := s.webhooks.ActiveForLinkOwner(ctx, *install.LinkID)
		if err == nil && len(webhooks) > 0 {
			payload := map[string]interface{}{
				"eventId":   eventID,
				"installId": in.InstallID,
				"linkId":    *install.LinkID,
				"eventName": in.EventName,
				"eventData": data,
				"timestamp": ts.Format(time.RFC3339Nano),
			}
			s.trigger.Trigger(ctx, webhooks, domain.WebhookEventConversion, eventID, payload)
			s.trigger.Trigger(ctx, webhooks, domain.WebhookEventSDK, eventID, payload)
		}
	}

	return &EventResult{EventID: eventID, Acknowledged: true}, nil
}

// AttributionResult mirrors GET /api/sdk/v1/attribution/:fingerprint.
type AttributionResult struct {
	Fingerprint string
	Attributed  bool
	Install     *domain.InstallEvent
	Click       *domain.ClickEvent
	Link        *domain.Link
}

func (s *SDKService) Attribution(ctx context.Context, fingerprint string) (*AttributionResult, error) {
	install, err := s.installs.LatestByFingerprintHash(ctx, fingerprint)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInstallNotFound
		}
		return nil, err
	}

	attributed := install.LinkID != nil

	var click *domain.ClickEvent
	if install.ClickID != nil {
		click, _ = s.clickEvents.GetByID(ctx, *install.ClickID)
	}

	var link *domain.Link
	if attributed {
		link, _ = s.links.GetByID(ctx, *install.LinkID, ports.LinkFilter{})
	}

	return &AttributionResult{
		Fingerprint: fingerprint,
		Attributed:  attributed,
		Install:     install,
		Click:       click,
		Link:        link,
	}, nil
}
