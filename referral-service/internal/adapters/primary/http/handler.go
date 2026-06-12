package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
	"go.uber.org/fx"
	"go.uber.org/zap"

	appconfig "github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
)

var Module = fx.Options(
	fx.Provide(NewHandler),
)

type Handler struct {
	svc ports.ReferralService
	log *zap.Logger
	cfg *appconfig.Config
}

func NewHandler(svc ports.ReferralService, log *zap.Logger, cfg *appconfig.Config) *Handler {
	return &Handler{svc: svc, log: log.Named("http-handler"), cfg: cfg}
}

func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		ref := v1.Group("/referral")
		ref.POST("/generate", h.GenerateLink)
		ref.GET("/link/:ref_code", h.GetLink)
		ref.GET("/user/:user_id/links", h.ListUserLinks)
		ref.POST("/event", h.TrackEvent)
		ref.POST("/confirm-install", h.ConfirmInstall)
		ref.POST("/activate", h.ActivateReferral)
		ref.GET("/status", h.GetReferralStatus)
		ref.GET("/user/:user_id/stats", h.GetUserStats)
		ref.GET("/qr/:ref_code", h.GetShareQRCode)
		ref.GET("/share/:ref_code", h.GetShareTargets)

		programs := v1.Group("/programs")
		programs.POST("", h.CreateRewardProgram)
		programs.GET("/active", h.GetActiveRewardProgram)
		programs.PATCH("/:program_id/status", h.SetRewardProgramStatus)
	}

	// Top-level landing route — matches the deep_link format produced by buildDeepLink.
	r.GET("/r/:ref_code", h.HandleShareRedirect)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// GenerateLink POST /api/v1/referral/generate
func (h *Handler) GenerateLink(c *gin.Context) {
	var req generateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, codeValidation, err.Error())
		return
	}

	channel := domain.Channel(req.Channel)
	if channel == "" {
		channel = domain.ChannelCopy
	} else if !validChannels[channel] {
		respBadRequest(c, codeInvalidField, "invalid channel")
		return
	}

	platform := req.Platform
	if platform == "" {
		platform = "universal"
	}

	resp, err := h.svc.GenerateLink(c.Request.Context(), ports.GenerateLinkRequest{
		OwnerUserID: req.OwnerUserID,
		Channel:     channel,
		Platform:    platform,
	})
	if err != nil {
		if errors.Is(err, domain.ErrRateLimitExceeded) {
			respTooManyRequests(c, codeRateLimit, err.Error())
			return
		}
		h.log.Error("generate link failed", zap.Error(err))
		respInternal(c)
		return
	}

	respCreated(c, generateLinkResponse{
		RefCode:   resp.RefCode,
		DeepLink:  resp.DeepLink,
		ExpiresAt: resp.ExpiresAt,
	})
}

// GetLink GET /api/v1/referral/link/:ref_code
func (h *Handler) GetLink(c *gin.Context) {
	refCode := c.Param("ref_code")

	link, err := h.svc.GetLink(c.Request.Context(), refCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respNotFound(c, "link not found")
			return
		}
		h.log.Error("get link failed", zap.Error(err))
		respInternal(c)
		return
	}

	respOK(c, link)
}

// ListUserLinks GET /api/v1/referral/user/:user_id/links?limit=20&cursor=...
func (h *Handler) ListUserLinks(c *gin.Context) {
	userID := c.Param("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	cursor := c.Query("cursor")

	links, nextCursor, err := h.svc.ListUserLinks(c.Request.Context(), userID, limit, cursor)
	if err != nil {
		h.log.Error("list user links failed", zap.Error(err))
		respInternal(c)
		return
	}

	respList(c, links, len(links), nextCursor)
}

// TrackEvent POST /api/v1/referral/event
func (h *Handler) TrackEvent(c *gin.Context) {
	var req trackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, codeValidation, err.Error())
		return
	}

	eventType := domain.EventType(req.EventType)
	if !validEventTypes[eventType] {
		respBadRequest(c, codeInvalidField, "invalid event_type")
		return
	}

	channel := domain.Channel(req.Channel)
	if channel != "" && !validChannels[channel] {
		respBadRequest(c, codeInvalidField, "invalid channel")
		return
	}

	if err := h.svc.TrackEvent(c.Request.Context(), ports.TrackEventRequest{
		RefCode:            req.RefCode,
		EventType:          eventType,
		Platform:           req.Platform,
		Channel:            channel,
		NewUserID:          req.NewUserID,
		IPAddress:          req.IPAddress,
		DeviceID:           req.DeviceID,
		UserAgent:          req.UserAgent,
		FirebaseInstanceID: req.FirebaseInstanceID,
	}); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respNotFound(c, "link not found")
			return
		}
		if errors.Is(err, domain.ErrLinkNotActive) {
			respUnprocessable(c, codeLinkNotActive, err.Error())
			return
		}
		h.log.Error("track event failed", zap.Error(err))
		respInternal(c)
		return
	}

	respOK(c, gin.H{"status": "tracked"})
}

// ConfirmInstall POST /api/v1/referral/confirm-install
func (h *Handler) ConfirmInstall(c *gin.Context) {
	var req confirmInstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, codeValidation, err.Error())
		return
	}

	resp, err := h.svc.ConfirmInstall(c.Request.Context(), ports.ConfirmInstallRequest{
		RefCode:            req.RefCode,
		NewUserID:          req.NewUserID,
		Platform:           req.Platform,
		DeviceID:           req.DeviceID,
		FirebaseInstanceID: req.FirebaseInstanceID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSelfReferral) {
			respBadRequest(c, codeSelfReferral, err.Error())
			return
		}
		if errors.Is(err, domain.ErrAlreadyReferred) {
			respBadRequest(c, codeAlreadyReferred, err.Error())
			return
		}
		h.log.Error("confirm install failed", zap.Error(err))
		respInternal(c)
		return
	}

	respOK(c, attributionResponse{
		Attributed:  resp.Attributed,
		OwnerUserID: resp.OwnerUserID,
		RewardType:  resp.RewardType,
	})
}

// ActivateReferral POST /api/v1/referral/activate
func (h *Handler) ActivateReferral(c *gin.Context) {
	var req activateReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, codeValidation, err.Error())
		return
	}

	resp, err := h.svc.ActivateReferral(c.Request.Context(), ports.ActivateReferralRequest{
		RefCode:            req.RefCode,
		UserID:             req.UserID,
		Platform:           req.Platform,
		DeviceID:           req.DeviceID,
		FirebaseInstanceID: req.FirebaseInstanceID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSelfReferral) {
			respBadRequest(c, codeSelfReferral, err.Error())
			return
		}
		if errors.Is(err, domain.ErrAlreadyReferred) {
			respBadRequest(c, codeAlreadyReferred, err.Error())
			return
		}
		h.log.Error("activate referral failed", zap.Error(err))
		respInternal(c)
		return
	}

	respOK(c, attributionResponse{
		Attributed:  resp.Attributed,
		OwnerUserID: resp.OwnerUserID,
		RewardType:  resp.RewardType,
	})
}

// GetReferralStatus GET /api/v1/referral/status?ref_code=ABC123
func (h *Handler) GetReferralStatus(c *gin.Context) {
	refCode := c.Query("ref_code")
	if refCode == "" {
		respBadRequest(c, codeValidation, "ref_code is required")
		return
	}

	status, err := h.svc.GetReferralStatus(c.Request.Context(), refCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respNotFound(c, "link not found")
			return
		}
		h.log.Error("get referral status failed", zap.Error(err))
		respInternal(c)
		return
	}

	respOK(c, status)
}

// GetUserStats GET /api/v1/referral/user/:user_id/stats
func (h *Handler) GetUserStats(c *gin.Context) {
	userID := c.Param("user_id")

	stats, err := h.svc.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("get stats failed", zap.Error(err))
		respInternal(c)
		return
	}

	respOK(c, stats)
}

// CreateRewardProgram POST /api/v1/programs
func (h *Handler) CreateRewardProgram(c *gin.Context) {
	var req createRewardProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, codeValidation, err.Error())
		return
	}

	tiers := make([]ports.RewardTierInput, len(req.Tiers))
	for i, t := range req.Tiers {
		maxCount := t.MaxCount
		if maxCount == 0 {
			maxCount = -1 // treat omitted max as unlimited
		}
		tiers[i] = ports.RewardTierInput{
			MinCount:    t.MinCount,
			MaxCount:    maxCount,
			RewardValue: t.RewardValue,
		}
	}

	program, err := h.svc.CreateRewardProgram(c.Request.Context(), ports.CreateRewardProgramRequest{
		Name:     req.Name,
		Tiers:    tiers,
		Activate: req.Activate,
	})
	if err != nil {
		h.log.Error("create reward program failed", zap.Error(err))
		respInternal(c)
		return
	}

	respCreated(c, program)
}

// GetActiveRewardProgram GET /api/v1/programs/active
func (h *Handler) GetActiveRewardProgram(c *gin.Context) {
	program, err := h.svc.GetActiveRewardProgram(c.Request.Context())
	if err != nil {
		h.log.Error("get active reward program failed", zap.Error(err))
		respInternal(c)
		return
	}
	if program == nil {
		respNotFound(c, "no active reward program")
		return
	}

	respOK(c, program)
}

// GetShareQRCode GET /api/v1/referral/qr/:ref_code
// Returns a PNG image of a QR code encoding the share link for the given ref code.
// Optional query params: size (64–1024, default 256), channel (tags the encoded URL for click attribution).
func (h *Handler) GetShareQRCode(c *gin.Context) {
	refCode := c.Param("ref_code")

	link, err := h.svc.GetLink(c.Request.Context(), refCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respNotFound(c, "link not found")
			return
		}
		h.log.Error("get link for qr failed", zap.Error(err))
		respInternal(c)
		return
	}
	if !link.IsActive() {
		respUnprocessable(c, codeLinkNotActive, "link is not active")
		return
	}

	size := 256
	if s := c.Query("size"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 64 && v <= 1024 {
			size = v
		}
	}

	channel := domain.Channel(c.Query("channel"))
	if channel != "" && !validChannels[channel] {
		respBadRequest(c, codeInvalidField, "invalid channel")
		return
	}

	shareURL := withChannel(link.DeepLink, channel)

	png, err := qrcode.Encode(shareURL, qrcode.Medium, size)
	if err != nil {
		h.log.Error("qr encode failed", zap.Error(err))
		respInternal(c)
		return
	}

	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "image/png", png)
}

// GetShareTargets GET /api/v1/referral/share/:ref_code
// Returns, for every supported channel, a tagged share URL and a matching QR image URL —
// lets a client build a multi-channel share sheet (Zalo/Facebook/WhatsApp/...) from one ref_code
// without calling /generate once per channel.
func (h *Handler) GetShareTargets(c *gin.Context) {
	refCode := c.Param("ref_code")

	link, err := h.svc.GetLink(c.Request.Context(), refCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respNotFound(c, "link not found")
			return
		}
		h.log.Error("get link for share targets failed", zap.Error(err))
		respInternal(c)
		return
	}
	if !link.IsActive() {
		respUnprocessable(c, codeLinkNotActive, "link is not active")
		return
	}

	targets := make([]shareTarget, 0, len(shareChannels))
	for _, sc := range shareChannels {
		shareURL := withChannel(link.DeepLink, sc.Channel)
		qrURL := fmt.Sprintf("%s/api/v1/referral/qr/%s?channel=%s", h.cfg.DeepLink.PublicURL, refCode, sc.Channel)
		targets = append(targets, shareTarget{
			Channel:  string(sc.Channel),
			Label:    sc.Label,
			ShareURL: shareURL,
			QRURL:    qrURL,
		})
	}

	respOK(c, targets)
}

// HandleShareRedirect GET /r/:ref_code
// The landing page behind every generated deep_link. Records the link_clicked event (the
// "invited at" timestamp), then routes the visitor to the right store:
//   - Android: redirects straight to Play Store with ref_code/channel embedded in the
//     `referrer` param, readable by the app via the Play Install Referrer API after install.
//   - iOS: Apple has no install-referrer equivalent, so this serves a tiny HTML page that
//     copies the ref_code to the clipboard before redirecting to the App Store. The app must
//     read the clipboard on first open and strip the "REF:" prefix to recover the ref_code.
//   - Other/unknown UA: redirects to the public site.
func (h *Handler) HandleShareRedirect(c *gin.Context) {
	refCode := c.Param("ref_code")
	channel := domain.Channel(c.Query("channel"))
	if !validChannels[channel] {
		channel = ""
	}

	link, err := h.svc.GetLink(c.Request.Context(), refCode)
	if err != nil || !link.IsActive() {
		c.Redirect(http.StatusFound, h.cfg.DeepLink.PublicURL)
		return
	}

	platform := detectPlatform(c.GetHeader("User-Agent"))

	if err := h.svc.TrackEvent(c.Request.Context(), ports.TrackEventRequest{
		RefCode:   refCode,
		EventType: domain.EventLinkClicked,
		Platform:  platform,
		Channel:   channel,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}); err != nil {
		h.log.Warn("share redirect: failed to track click", zap.String("ref_code", refCode), zap.Error(err))
	}

	switch platform {
	case "android":
		referrer := fmt.Sprintf("ref_code=%s&channel=%s", refCode, channel)
		sep := "?"
		if strings.Contains(h.cfg.DeepLink.PlayStoreURL, "?") {
			sep = "&"
		}
		c.Redirect(http.StatusFound, h.cfg.DeepLink.PlayStoreURL+sep+"referrer="+url.QueryEscape(referrer))
	case "ios":
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "text/html; charset=utf-8", buildClipboardRedirectHTML(refCode, h.cfg.DeepLink.AppStoreURL))
	default:
		c.Redirect(http.StatusFound, h.cfg.DeepLink.PublicURL)
	}
}

// withChannel appends a channel query param to a deep link, if channel is non-empty.
func withChannel(deepLink string, channel domain.Channel) string {
	if channel == "" {
		return deepLink
	}
	sep := "?"
	if strings.Contains(deepLink, "?") {
		sep = "&"
	}
	return deepLink + sep + "channel=" + url.QueryEscape(string(channel))
}

// detectPlatform sniffs the User-Agent header for a coarse "ios" | "android" | "universal" classification.
func detectPlatform(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipad"), strings.Contains(ua, "ipod"):
		return "ios"
	case strings.Contains(ua, "android"):
		return "android"
	default:
		return "universal"
	}
}

// buildClipboardRedirectHTML renders a minimal landing page that copies the ref_code marker to
// the clipboard before redirecting to the App Store — the deferred deep-link workaround for iOS.
func buildClipboardRedirectHTML(refCode, appStoreURL string) []byte {
	return []byte(fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Opening app store…</title></head>
<body>
<script>
  (function () {
    var marker = "REF:%s";
    var dest = %q;
    function go() { window.location.replace(dest); }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(marker).then(go, go);
    } else {
      go();
    }
    setTimeout(go, 800);
  })();
</script>
</body></html>`, refCode, appStoreURL))
}

// SetRewardProgramStatus PATCH /api/v1/programs/:program_id/status
func (h *Handler) SetRewardProgramStatus(c *gin.Context) {
	programID := c.Param("program_id")
	var req setRewardProgramStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respBadRequest(c, codeValidation, err.Error())
		return
	}

	if err := h.svc.SetRewardProgramStatus(c.Request.Context(), programID, req.Active); err != nil {
		h.log.Error("set reward program status failed", zap.Error(err))
		respInternal(c)
		return
	}

	respOK(c, gin.H{"status": "updated"})
}
