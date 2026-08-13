package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/JIeeiroSst/shortlink-service/internal/adapters/repo"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"github.com/google/uuid"
)

type RedirectInput struct {
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

	AbuseReportURL string
}

// RedirectOutcomeKind discriminates RedirectOutcome.
type RedirectOutcomeKind int

const (
	OutcomeNotFound RedirectOutcomeKind = iota
	OutcomeNoDestination
	OutcomeWarnHTML
	OutcomeInterstitialHTML
	OutcomeRedirect
)

// RedirectOutcome mirrors the possible responses handleRedirect() sends.
type RedirectOutcome struct {
	Kind        RedirectOutcomeKind
	HTML        string
	RedirectURL string
}

type RedirectService struct {
	links        ports.LinkRepository
	cache        ports.Cache
	clickTracker *ClickTracker
	geoip        ports.GeoIPLookup
}

func NewRedirectService(links ports.LinkRepository, cache ports.Cache, clickTracker *ClickTracker, geoip ports.GeoIPLookup) *RedirectService {
	return &RedirectService{links, cache, clickTracker, geoip}
}

func (s *RedirectService) Resolve(ctx context.Context, in RedirectInput) (*RedirectOutcome, error) {
	link, err := resolveLinkCached(ctx, s.links, s.cache, in.ShortCode, in.TemplateSlug)
	if err != nil {
		if errors.Is(err, ErrLinkNotFound) || errors.Is(err, repo.ErrNotFound) {
			return &RedirectOutcome{Kind: OutcomeNotFound}, nil
		}
		return nil, err
	}

	safety := domain.EvaluateLinkSafety(domain.LinkSafetyInput{
		IsActive:         &link.IsActive,
		WarnAt:           link.WarnAt,
		OwnerSuspendedAt: link.OwnerSuspendedAt,
	})
	if safety == domain.SafetyBlock {
		return &RedirectOutcome{Kind: OutcomeNotFound}, nil
	}
	if safety == domain.SafetyWarn {
		html := domain.GenerateWarningLinkHTML(firstNonEmpty(link.OriginalURL, derefOrEmpty(link.WebFallbackURL), derefOrEmpty(link.DeepLinkPath)), in.AbuseReportURL)
		return &RedirectOutcome{Kind: OutcomeWarnHTML, HTML: html}, nil
	}

	device := domain.DetectDevice(in.UserAgent)
	countryCode := ""
	if len(link.TargetingRules.Countries) > 0 {
		if geo := s.geoip.Lookup(in.IP); geo.CountryCode != nil {
			countryCode = *geo.CountryCode
		}
	}
	targetingCtx := domain.TargetingContext{
		Device:          device,
		CountryCode:     countryCode,
		PrimaryLanguage: domain.PrimaryLanguageFromAcceptLanguage(in.AcceptLanguage),
	}
	if !domain.EvaluateTargeting(link.TargetingRules, targetingCtx) {
		return &RedirectOutcome{Kind: OutcomeNotFound}, nil
	}

	clickID := uuid.NewString()
	s.clickTracker.TrackAsync(ClickTrackInput{
		Link: link, IP: in.IP, UserAgent: in.UserAgent, Referrer: in.Referrer,
		AcceptLanguage: in.AcceptLanguage, Method: in.Method,
		TrustEdgeBotHeader: in.TrustEdgeBotHeader, EdgeBotHeaderValue: in.EdgeBotHeaderValue,
		UTMSource: in.UTMSource, UTMMedium: in.UTMMedium, UTMCampaign: in.UTMCampaign,
		FPTimezone: in.FPTimezone, FPLanguage: in.FPLanguage,
		FPScreenWidth: in.FPScreenWidth, FPScreenHeight: in.FPScreenHeight,
		PreGeneratedClickID:        clickID,
		IncludeUTMInWebhookPayload: true,
	})

	templateSettings := domain.LinkTemplateSettings{}
	if link.TemplateSettings != nil {
		templateSettings = *link.TemplateSettings
	}
	orgAppConfig := domain.OrganizationAppConfig{}
	if link.OrgSettings != nil {
		orgAppConfig = link.OrgSettings.AppConfig
	}

	iosURL := firstNonEmpty(derefOrEmpty(link.IOSAppStoreURL), templateSettings.DefaultIosURL, orgAppConfig.IosAppStoreURL)
	androidURL := firstNonEmpty(derefOrEmpty(link.AndroidAppStoreURL), templateSettings.DefaultAndroidURL, orgAppConfig.AndroidAppStoreURL)
	webFallbackURL := firstNonEmpty(derefOrEmpty(link.WebFallbackURL), templateSettings.DefaultWebFallbackURL, orgAppConfig.WebFallbackURL)

	redirectURL := link.OriginalURL
	useSchemeURL := false

	switch device {
	case domain.DeviceIOS:
		switch {
		case derefOrEmpty(link.IOSUniversalLink) != "":
			redirectURL = *link.IOSUniversalLink
		case derefOrEmpty(link.AppScheme) != "" && derefOrEmpty(link.DeepLinkPath) != "":
			redirectURL = buildSchemeURL(*link.AppScheme, *link.DeepLinkPath)
			useSchemeURL = true
		default:
			if fb := domain.PickMobileFallbackURL(domain.DeviceIOS, in.UserAgent, iosURL, androidURL, webFallbackURL); fb != nil {
				redirectURL = fb.URL
			}
		}
	case domain.DeviceAndroid:
		switch {
		case derefOrEmpty(link.AndroidAppLink) != "":
			redirectURL = *link.AndroidAppLink
		case derefOrEmpty(link.AppScheme) != "" && derefOrEmpty(link.DeepLinkPath) != "":
			redirectURL = buildSchemeURL(*link.AppScheme, *link.DeepLinkPath)
			useSchemeURL = true
		default:
			if fb := domain.PickMobileFallbackURL(domain.DeviceAndroid, in.UserAgent, iosURL, androidURL, webFallbackURL); fb != nil {
				redirectURL = fb.URL
			}
		}
	default: // web
		if webFallbackURL != "" {
			redirectURL = webFallbackURL
		} else {
			redirectURL = link.OriginalURL
		}
	}

	if redirectURL == "" {
		return &RedirectOutcome{Kind: OutcomeNoDestination}, nil
	}

	finalURL := redirectURL
	if !useSchemeURL {
		finalURL = domain.BuildRedirectURL(redirectURL, link.UTMParameters)
		if finalURL == "" {
			finalURL = redirectURL
		}
		if len(link.DeepLinkParameters) > 0 {
			if withParams, err := appendQueryParams(finalURL, link.DeepLinkParameters); err == nil {
				finalURL = withParams
			}
		}
		if link.AppendClickID {
			if withClick, err := appendQueryParams(finalURL, map[string]interface{}{"lf_click": clickID}); err == nil {
				finalURL = withClick
			}
		}
	} else if len(link.DeepLinkParameters) > 0 {
		finalURL += "?" + encodeQueryParams(link.DeepLinkParameters)
	}

	if (device == domain.DeviceIOS || device == domain.DeviceAndroid) && derefOrEmpty(link.AppScheme) != "" {
		deepPath := strings.TrimPrefix(derefOrEmpty(link.DeepLinkPath), "/")
		schemeURL := *link.AppScheme + "://" + deepPath

		fb := domain.PickMobileFallbackURL(device, in.UserAgent, iosURL, androidURL, webFallbackURL)
		storeFallback := link.OriginalURL
		if fb != nil {
			storeFallback = fb.URL
		}

		if storeFallback != "" {
			fullSchemeURL := schemeURL
			if len(link.DeepLinkParameters) > 0 {
				sep := "?"
				if strings.Contains(fullSchemeURL, "?") {
					sep = "&"
				}
				fullSchemeURL += sep + encodeQueryParams(link.DeepLinkParameters)
			}
			title := derefOrEmpty(link.Title)
			if title == "" {
				title = derefOrEmpty(link.OGTitle)
			}
			html := generateInterstitialHTML(fullSchemeURL, storeFallback, title)
			return &RedirectOutcome{Kind: OutcomeInterstitialHTML, HTML: html}, nil
		}
	}

	return &RedirectOutcome{Kind: OutcomeRedirect, RedirectURL: finalURL}, nil
}

func buildSchemeURL(appScheme, deepLinkPath string) string {
	return appScheme + "://" + strings.TrimPrefix(deepLinkPath, "/")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func appendQueryParams(rawURL string, params map[string]interface{}) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, fmt.Sprintf("%v", v))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func encodeQueryParams(params map[string]interface{}) string {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, fmt.Sprintf("%v", v))
	}
	return q.Encode()
}

func generateInterstitialHTML(schemeURL, fallbackURL, title string) string {
	safeSchemeURL := strings.ReplaceAll(strings.ReplaceAll(schemeURL, `"`, "&quot;"), "<", "&lt;")
	safeFallbackURL := strings.ReplaceAll(strings.ReplaceAll(fallbackURL, `"`, "&quot;"), "<", "&lt;")
	safeTitle := title
	if safeTitle == "" {
		safeTitle = "the app"
	}
	safeTitle = strings.ReplaceAll(strings.ReplaceAll(safeTitle, "<", "&lt;"), ">", "&gt;")

	return `<!DOCTYPE html>
<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Opening ` + safeTitle + `...</title>
<style>
  body { font-family: -apple-system, system-ui, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; background: #f9fafb; color: #111827; text-align: center; }
  .container { padding: 2rem; }
  .spinner { width: 40px; height: 40px; border: 3px solid #e5e7eb; border-top-color: #3b82f6; border-radius: 50%; animation: spin 0.8s linear infinite; margin: 0 auto 1.5rem; }
  @keyframes spin { to { transform: rotate(360deg); } }
  h1 { font-size: 1.25rem; font-weight: 600; margin: 0 0 0.5rem; }
  p { font-size: 0.875rem; color: #6b7280; margin: 0 0 2rem; }
  .btn { display: inline-block; padding: 0.75rem 1.5rem; border-radius: 0.5rem; font-size: 0.875rem; font-weight: 500; text-decoration: none; margin: 0.25rem; }
  .btn-primary { background: #3b82f6; color: #fff; }
  .btn-secondary { background: #e5e7eb; color: #374151; }
</style>
</head><body>
<div class="container">
  <div class="spinner"></div>
  <h1>Opening ` + safeTitle + `...</h1>
  <p>If the app doesn't open automatically:</p>
  <a class="btn btn-primary" id="open-btn" href="` + safeSchemeURL + `">Open App</a>
  <a class="btn btn-secondary" id="store-btn" href="` + safeFallbackURL + `">Download App</a>
</div>
<script>
  var hash = window.location.hash || '';
  var schemeUrl = "` + safeSchemeURL + `" + hash;
  document.getElementById('open-btn').href = schemeUrl;
  window.location = schemeUrl;
  setTimeout(function() { window.location.replace("` + safeFallbackURL + `"); }, 1500);
</script>
</body></html>`
}
