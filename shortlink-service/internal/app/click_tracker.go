package app

import (
	"context"
	"strconv"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"go.uber.org/zap"
)

type ClickTrackInput struct {
	Link                              *domain.Link
	IP                                string
	UserAgent                         string
	Referrer                          string
	AcceptLanguage                    string
	Method                            string
	TrustEdgeBotHeader                bool
	EdgeBotHeaderValue                string
	UTMSource, UTMMedium, UTMCampaign string
	FPTimezone                        string
	FPLanguage                        string
	FPScreenWidth                     string
	FPScreenHeight                    string
	PreGeneratedClickID               string
	IncludeUTMInWebhookPayload        bool
}

type ClickTracker struct {
	clicks       ports.ClickEventRepository
	fingerprints ports.FingerprintRepository
	webhooks     ports.WebhookRepository
	trigger      *WebhookTrigger
	geoip        ports.GeoIPLookup
	log          *zap.Logger
}

func NewClickTracker(
	clicks ports.ClickEventRepository,
	fingerprints ports.FingerprintRepository,
	webhooks ports.WebhookRepository,
	trigger *WebhookTrigger,
	geoip ports.GeoIPLookup,
	log *zap.Logger,
) *ClickTracker {
	return &ClickTracker{clicks, fingerprints, webhooks, trigger, geoip, log}
}

func (t *ClickTracker) TrackAsync(in ClickTrackInput) {
	go func() {
		ctx := context.Background()
		if err := t.track(ctx, in); err != nil {
			t.log.Error("click tracking failed", zap.Error(err), zap.String("linkId", in.Link.ID))
		}
	}()
}

func (t *ClickTracker) track(ctx context.Context, in ClickTrackInput) error {
	link := in.Link
	deviceType := domain.DetectDevice(in.UserAgent)
	parsedUA := domain.ParseUserAgent(in.UserAgent)
	geo := t.geoip.Lookup(in.IP)

	edgeSignal := domain.EdgeBotSignal(in.TrustEdgeBotHeader, in.EdgeBotHeaderValue)
	botClass := domain.ClassifyBot(in.UserAgent, in.Method, edgeSignal)

	fpTimezone := in.FPTimezone
	if fpTimezone == "" && geo.Timezone != nil {
		fpTimezone = *geo.Timezone
	}
	fpLanguage := in.FPLanguage
	if fpLanguage == "" {
		fpLanguage = domain.FingerprintLanguageFromAcceptLanguage(in.AcceptLanguage)
	}
	var fpScreenWidth, fpScreenHeight *int
	if v, err := strconv.Atoi(in.FPScreenWidth); err == nil {
		fpScreenWidth = &v
	}
	if v, err := strconv.Atoi(in.FPScreenHeight); err == nil {
		fpScreenHeight = &v
	}

	platformStr := string(parsedUA.Platform)
	event := &domain.ClickEvent{
		LinkID:      link.ID,
		IPAddress:   strPtr(in.IP),
		UserAgent:   strPtr(in.UserAgent),
		DeviceType:  strPtr(string(deviceType)),
		Platform:    strPtr(platformStr),
		CountryCode: geo.CountryCode,
		CountryName: geo.CountryName,
		Region:      geo.Region,
		City:        geo.City,
		Latitude:    geo.Latitude,
		Longitude:   geo.Longitude,
		Timezone:    geo.Timezone,
		UTMSource:   strPtr(in.UTMSource),
		UTMMedium:   strPtr(in.UTMMedium),
		UTMCampaign: strPtr(in.UTMCampaign),
		Referrer:    strPtr(in.Referrer),
		IsBot:       botClass.IsBot,
		BotReason:   botReasonPtr(botClass),
	}

	var clickID string
	if in.PreGeneratedClickID != "" {
		if err := t.clicks.InsertWithID(ctx, in.PreGeneratedClickID, event); err != nil {
			return err
		}
		clickID = in.PreGeneratedClickID
	} else {
		if err := t.clicks.Insert(ctx, event); err != nil {
			return err
		}
		clickID = event.ID
	}

	fp := domain.FingerprintData{
		IPAddress:       in.IP,
		UserAgent:       in.UserAgent,
		Timezone:        fpTimezone,
		Language:        fpLanguage,
		ScreenWidth:     fpScreenWidth,
		ScreenHeight:    fpScreenHeight,
		Platform:        string(deviceType),
		PlatformVersion: parsedUA.PlatformVersion,
	}
	if err := t.fingerprints.StoreForClick(ctx, clickID, fp); err != nil {
		return err
	}

	if link.UserID != nil {
		webhooks, err := t.webhooks.ActiveForUser(ctx, *link.UserID)
		if err == nil && len(webhooks) > 0 {
			payload := map[string]interface{}{
				"id":          clickID,
				"linkId":      link.ID,
				"clickedAt":   time.Now().UTC().Format(time.RFC3339Nano),
				"ipAddress":   in.IP,
				"userAgent":   in.UserAgent,
				"deviceType":  deviceType,
				"platform":    platformStr,
				"countryCode": geo.CountryCode,
				"countryName": geo.CountryName,
				"region":      geo.Region,
				"city":        geo.City,
				"latitude":    geo.Latitude,
				"longitude":   geo.Longitude,
				"timezone":    geo.Timezone,
				"referrer":    in.Referrer,
			}
			if in.IncludeUTMInWebhookPayload {
				payload["utmSource"] = in.UTMSource
				payload["utmMedium"] = in.UTMMedium
				payload["utmCampaign"] = in.UTMCampaign
			}
			t.trigger.Trigger(ctx, webhooks, domain.WebhookEventClick, clickID, payload)
		}
	}

	return nil
}

func botReasonPtr(c domain.BotClassification) *string {
	if c.Reason == "" {
		return nil
	}
	s := string(c.Reason)
	return &s
}
