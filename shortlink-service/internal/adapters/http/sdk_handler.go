package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type SDKHandler struct {
	sdk                *app.SDKService
	trustProxy         domain.TrustProxy
	trustEdgeBotHeader bool
}

func NewSDKHandler(sdk *app.SDKService, trustProxy domain.TrustProxy, trustEdgeBotHeader bool) *SDKHandler {
	return &SDKHandler{sdk, trustProxy, trustEdgeBotHeader}
}

// --- POST /api/sdk/v1/install ---

type installRequest struct {
	IPAddress              *string `json:"ipAddress"`
	UserAgent              string  `json:"userAgent" binding:"required"`
	Timezone               *string `json:"timezone"`
	Language               *string `json:"language"`
	ScreenWidth            *int    `json:"screenWidth"`
	ScreenHeight           *int    `json:"screenHeight"`
	Platform               *string `json:"platform"`
	PlatformVersion        *string `json:"platformVersion"`
	DeviceID               *string `json:"deviceId"`
	AttributionWindowHours *int    `json:"attributionWindowHours"`
	SDKName                *string `json:"sdkName"`
	SDKVersion             *string `json:"sdkVersion"`
	AppToken               *string `json:"appToken"`
}

type installResponse struct {
	InstallID        string                 `json:"installId"`
	Attributed       bool                   `json:"attributed"`
	ConfidenceScore  int                    `json:"confidenceScore"`
	MatchedFactors   []string               `json:"matchedFactors"`
	DeepLinkData     map[string]interface{} `json:"deepLinkData"`
	ClientReportedIP *string                `json:"clientReportedIp,omitempty"`
}

func (h *SDKHandler) Install(c *gin.Context) {
	var req installRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	in := app.InstallInput{
		TrustedIP:              ClientIP(c, h.trustProxy),
		UserAgent:              req.UserAgent,
		Timezone:               derefStr(req.Timezone),
		Language:               derefStr(req.Language),
		ScreenWidth:            req.ScreenWidth,
		ScreenHeight:           req.ScreenHeight,
		Platform:               derefStr(req.Platform),
		PlatformVersion:        derefStr(req.PlatformVersion),
		DeviceID:               derefStr(req.DeviceID),
		AttributionWindowHours: derefInt(req.AttributionWindowHours),
		SDKName:                derefStr(req.SDKName),
		SDKVersion:             derefStr(req.SDKVersion),
	}

	result, err := h.sdk.Install(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record install event", "message": err.Error()})
		return
	}

	resp := installResponse{
		InstallID: result.InstallID, Attributed: result.Attributed,
		ConfidenceScore: result.ConfidenceScore, MatchedFactors: result.MatchedFactors,
		DeepLinkData: result.DeepLinkData,
	}
	if req.IPAddress != nil {
		resp.ClientReportedIP = req.IPAddress
	}
	c.JSON(http.StatusOK, resp)
}

// --- GET /api/sdk/v1/resolve/:shortCode and /:templateSlug/:shortCode ---

type resolveResponse struct {
	ShortCode        string                 `json:"shortCode"`
	LinkID           string                 `json:"linkId"`
	DeepLinkPath     *string                `json:"deepLinkPath,omitempty"`
	AppScheme        *string                `json:"appScheme,omitempty"`
	IOSUrl           *string                `json:"iosUrl,omitempty"`
	AndroidUrl       *string                `json:"androidUrl,omitempty"`
	WebUrl           *string                `json:"webUrl,omitempty"`
	UTMParameters    domain.UTMParameters   `json:"utmParameters"`
	CustomParameters map[string]interface{} `json:"customParameters"`
	ClickedAt        time.Time              `json:"clickedAt"`
}

func (h *SDKHandler) resolve(c *gin.Context, shortCode, templateSlug string) {
	in := app.ResolveInput{
		ShortCode: shortCode, TemplateSlug: templateSlug,
		IP: ClientIP(c, h.trustProxy), UserAgent: c.Request.UserAgent(),
		Referrer: c.GetHeader("Referer"), AcceptLanguage: c.GetHeader("Accept-Language"),
		Method:             c.Request.Method,
		TrustEdgeBotHeader: h.trustEdgeBotHeader, EdgeBotHeaderValue: c.GetHeader("X-LF-Bot"),
		UTMSource: c.Query("utm_source"), UTMMedium: c.Query("utm_medium"), UTMCampaign: c.Query("utm_campaign"),
		FPTimezone: c.Query("fp_tz"), FPLanguage: c.Query("fp_lang"),
		FPScreenWidth: c.Query("fp_sw"), FPScreenHeight: c.Query("fp_sh"),
	}

	result, err := h.sdk.Resolve(c.Request.Context(), in)
	if err != nil {
		if errors.Is(err, app.ErrLinkNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve link", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resolveResponse{
		ShortCode: result.ShortCode, LinkID: result.LinkID, DeepLinkPath: result.DeepLinkPath,
		AppScheme: result.AppScheme, IOSUrl: result.IOSUrl, AndroidUrl: result.AndroidUrl, WebUrl: result.WebUrl,
		UTMParameters: result.UTMParameters, CustomParameters: result.CustomParameters, ClickedAt: result.ClickedAt,
	})
}

func (h *SDKHandler) ResolveByShortCode(c *gin.Context) {
	h.resolve(c, c.Param("shortCode"), "")
}

func (h *SDKHandler) ResolveByTemplateAndShortCode(c *gin.Context) {
	h.resolve(c, c.Param("shortCode"), c.Param("templateSlug"))
}

// --- POST /api/sdk/v1/event ---

type eventRequest struct {
	InstallID         string                 `json:"installId" binding:"required,uuid"`
	EventName         string                 `json:"eventName" binding:"required"`
	EventData         map[string]interface{} `json:"eventData"`
	Timestamp         *string                `json:"timestamp"`
	AttributedLinkID  *string                `json:"attributedLinkId"`
	AttributedClickID *string                `json:"attributedClickId"`
	LinkOpenedAt      *string                `json:"linkOpenedAt"`
	SessionID         *string                `json:"sessionId"`
	SDKName           *string                `json:"sdkName"`
	SDKVersion        *string                `json:"sdkVersion"`
}

type eventResponse struct {
	EventID      string `json:"eventId"`
	Acknowledged bool   `json:"acknowledged"`
}

func (h *SDKHandler) Event(c *gin.Context) {
	var req eventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	in := app.EventInput{
		InstallID: req.InstallID, EventName: req.EventName, EventData: req.EventData,
		AttributedLinkID: derefStr(req.AttributedLinkID), AttributedClickID: derefStr(req.AttributedClickID),
		SessionID: derefStr(req.SessionID), SDKName: derefStr(req.SDKName), SDKVersion: derefStr(req.SDKVersion),
	}
	if req.Timestamp != nil {
		if t, err := time.Parse(time.RFC3339, *req.Timestamp); err == nil {
			in.Timestamp = &t
		}
	}
	if req.LinkOpenedAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.LinkOpenedAt); err == nil {
			in.LinkOpenedAt = &t
		}
	}

	result, err := h.sdk.Event(c.Request.Context(), in)
	if err != nil {
		if errors.Is(err, app.ErrInstallNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Install event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track event", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, eventResponse{EventID: result.EventID, Acknowledged: result.Acknowledged})
}

// --- GET /api/sdk/v1/attribution/:fingerprint ---

type installEventDTO struct {
	ID                string     `json:"id"`
	InstalledAt       time.Time  `json:"installedAt"`
	FirstOpenAt       *time.Time `json:"firstOpenAt"`
	ConfidenceScore   float64    `json:"confidenceScore"`
	DeepLinkRetrieved bool       `json:"deepLinkRetrieved"`
}

type clickEventDTO struct {
	ID          string  `json:"id"`
	ClickedAt   string  `json:"clickedAt"`
	DeviceType  *string `json:"deviceType"`
	Platform    *string `json:"platform"`
	CountryCode *string `json:"countryCode"`
	City        *string `json:"city"`
}

type attributionLinkDataDTO struct {
	ShortCode          string                 `json:"shortCode"`
	OriginalURL        string                 `json:"originalUrl"`
	IOSUrl             *string                `json:"iosUrl"`
	AndroidUrl         *string                `json:"androidUrl"`
	WebFallbackURL     *string                `json:"webFallbackUrl"`
	UTMParameters      domain.UTMParameters   `json:"utmParameters"`
	DeepLinkParameters map[string]interface{} `json:"deepLinkParameters"`
}

type attributionResponse struct {
	Fingerprint  string                  `json:"fingerprint"`
	Attributed   bool                    `json:"attributed"`
	InstallEvent installEventDTO         `json:"installEvent"`
	ClickEvent   *clickEventDTO          `json:"clickEvent"`
	LinkData     *attributionLinkDataDTO `json:"linkData"`
}

func (h *SDKHandler) Attribution(c *gin.Context) {
	fingerprint := c.Param("fingerprint")

	result, err := h.sdk.Attribution(c.Request.Context(), fingerprint)
	if err != nil {
		if errors.Is(err, app.ErrInstallNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No install event found for this fingerprint"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attribution data", "message": err.Error()})
		return
	}

	resp := attributionResponse{
		Fingerprint: result.Fingerprint,
		Attributed:  result.Attributed,
		InstallEvent: installEventDTO{
			ID: result.Install.ID, InstalledAt: result.Install.InstalledAt, FirstOpenAt: result.Install.FirstOpenAt,
			ConfidenceScore: derefFloat(result.Install.ConfidenceScore), DeepLinkRetrieved: result.Install.DeepLinkRetrieved,
		},
	}
	if result.Click != nil {
		resp.ClickEvent = &clickEventDTO{
			ID: result.Click.ID, ClickedAt: result.Click.ClickedAt.Format(time.RFC3339Nano),
			DeviceType: result.Click.DeviceType, Platform: result.Click.Platform,
			CountryCode: result.Click.CountryCode, City: result.Click.City,
		}
	}
	if result.Attributed && result.Link != nil {
		resp.LinkData = &attributionLinkDataDTO{
			ShortCode: result.Link.ShortCode, OriginalURL: result.Link.OriginalURL,
			IOSUrl: result.Link.IOSAppStoreURL, AndroidUrl: result.Link.AndroidAppStoreURL,
			WebFallbackURL: result.Link.WebFallbackURL, UTMParameters: result.Link.UTMParameters,
			DeepLinkParameters: result.Link.DeepLinkParameters,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// --- GET /api/sdk/v1/health ---

func (h *SDKHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"version":   "v1",
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func derefFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
