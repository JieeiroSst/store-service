package http

import (
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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

var redirectTmpl = template.Must(template.New("redirect").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Opening App...</title>
  <style>body{font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;flex-direction:column;gap:12px;}</style>
</head>
<body>
  <p>Opening the app...</p>
  <a href="{{.StoreURL}}">Click here if the app doesn't open</a>
  <script>
    (function() {
      var storeURL = {{.StoreURL}};
      var deepLink = {{.DeepLink}};
      var start = Date.now();
      if (deepLink) {
        window.location.href = deepLink;
        setTimeout(function() {
          if (!document.hidden && Date.now() - start < 2500) {
            window.location.href = storeURL;
          }
        }, 1500);
      } else {
        window.location.href = storeURL;
      }
    })();
  </script>
</body>
</html>`))

type redirectData struct {
	StoreURL string
	DeepLink string
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/r/:ref_code", h.Redirect)

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
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// Redirect GET /r/:ref_code
// Serves an HTML page that tries to open the app via deep link,
// then falls back to the appropriate store URL based on User-Agent.
func (h *Handler) Redirect(c *gin.Context) {
	refCode := c.Param("ref_code")

	link, err := h.svc.GetLink(c.Request.Context(), refCode)
	if err != nil || !link.IsActive() {
		c.Redirect(http.StatusFound, h.cfg.DeepLink.AppStoreURL)
		return
	}

	ua := c.GetHeader("User-Agent")
	platform := detectPlatform(ua)

	_ = h.svc.TrackEvent(c.Request.Context(), ports.TrackEventRequest{
		RefCode:   refCode,
		EventType: domain.EventLinkClicked,
		Platform:  platform,
		IPAddress: c.ClientIP(),
		UserAgent: ua,
	})

	storeURL := h.storeURL(refCode, platform)
	deepLink := ""
	if h.cfg.DeepLink.AppURLScheme != "" {
		deepLink = h.cfg.DeepLink.AppURLScheme + "?ref=" + url.QueryEscape(refCode)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = redirectTmpl.Execute(c.Writer, redirectData{
		StoreURL: storeURL,
		DeepLink: deepLink,
	})
}

func (h *Handler) storeURL(refCode, platform string) string {
	switch platform {
	case "ios":
		return h.cfg.DeepLink.AppStoreURL
	case "android":
		return h.cfg.DeepLink.PlayStoreURL + "&referrer=" + url.QueryEscape("ref_code="+refCode)
	default:
		return h.cfg.DeepLink.AppStoreURL
	}
}

func detectPlatform(ua string) string {
	ua = strings.ToLower(ua)
	if strings.ContainsAny(ua, "iphone ipad ipod") {
		return "ios"
	}
	if strings.Contains(ua, "android") {
		return "android"
	}
	return "unknown"
}


type generateLinkRequest struct {
	OwnerUserID string `json:"owner_user_id" binding:"required"`
	Channel     string `json:"channel"` // "copy" | "whatsapp" | "facebook"
	Platform    string `json:"platform"` // "ios" | "android" | "universal"
}

type trackEventRequest struct {
	RefCode   string `json:"ref_code"   binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	Platform  string `json:"platform"`
	NewUserID string `json:"new_user_id"`
	IPAddress string `json:"ip_address"`
	DeviceID  string `json:"device_id"`
	UserAgent string `json:"user_agent"`
}

type confirmInstallRequest struct {
	RefCode   string `json:"ref_code"   binding:"required"`
	NewUserID string `json:"new_user_id" binding:"required"`
	Platform  string `json:"platform"`
	DeviceID  string `json:"device_id"`
}

var validChannels = map[domain.Channel]bool{
	domain.ChannelCopy:      true,
	domain.ChannelWhatsApp:  true,
	domain.ChannelFacebook:  true,
	domain.ChannelInstagram: true,
	domain.ChannelOther:     true,
}

var validEventTypes = map[domain.EventType]bool{
	domain.EventLinkCopied:   true,
	domain.EventLinkClicked:  true,
	domain.EventAppInstalled: true,
	domain.EventRegistered:   true,
	domain.EventRewardGiven:  true,
}

// GenerateLink POST /api/v1/referral/generate
func (h *Handler) GenerateLink(c *gin.Context) {
	var req generateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	channel := domain.Channel(req.Channel)
	if channel == "" {
		channel = domain.ChannelCopy
	} else if !validChannels[channel] {
		c.JSON(http.StatusBadRequest, errorResponse("invalid channel"))
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
		h.log.Error("generate link failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ref_code":   resp.RefCode,
		"deep_link":  resp.DeepLink,
		"expires_at": resp.ExpiresAt,
	})
}

// GetLink GET /api/v1/referral/link/:ref_code
func (h *Handler) GetLink(c *gin.Context) {
	refCode := c.Param("ref_code")
	link, err := h.svc.GetLink(c.Request.Context(), refCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorResponse("link not found"))
			return
		}
		h.log.Error("get link failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, link)
}

// ListUserLinks GET /api/v1/referral/user/:user_id/links?limit=20&cursor=...
func (h *Handler) ListUserLinks(c *gin.Context) {
	userID := c.Param("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	cursor := c.Query("cursor")

	links, nextCursor, err := h.svc.ListUserLinks(c.Request.Context(), userID, limit, cursor)
	if err != nil {
		h.log.Error("list user links failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       links,
		"next_cursor": nextCursor,
		"count":       len(links),
	})
}

// TrackEvent POST /api/v1/referral/event
func (h *Handler) TrackEvent(c *gin.Context) {
	var req trackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	eventType := domain.EventType(req.EventType)
	if !validEventTypes[eventType] {
		c.JSON(http.StatusBadRequest, errorResponse("invalid event_type"))
		return
	}

	if err := h.svc.TrackEvent(c.Request.Context(), ports.TrackEventRequest{
		RefCode:   req.RefCode,
		EventType: eventType,
		Platform:  req.Platform,
		NewUserID: req.NewUserID,
		IPAddress: req.IPAddress,
		DeviceID:  req.DeviceID,
		UserAgent: req.UserAgent,
	}); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorResponse("link not found"))
			return
		}
		if errors.Is(err, domain.ErrLinkNotActive) {
			c.JSON(http.StatusUnprocessableEntity, errorResponse(err.Error()))
			return
		}
		h.log.Error("track event failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "tracked"})
}

// ConfirmInstall POST /api/v1/referral/confirm-install
func (h *Handler) ConfirmInstall(c *gin.Context) {
	var req confirmInstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	resp, err := h.svc.ConfirmInstall(c.Request.Context(), ports.ConfirmInstallRequest{
		RefCode:   req.RefCode,
		NewUserID: req.NewUserID,
		Platform:  req.Platform,
		DeviceID:  req.DeviceID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSelfReferral) {
			c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
			return
		}
		h.log.Error("confirm install failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"attributed":    resp.Attributed,
		"owner_user_id": resp.OwnerUserID,
		"reward_type":   resp.RewardType,
	})
}

// GetUserStats GET /api/v1/referral/user/:user_id/stats
func (h *Handler) GetUserStats(c *gin.Context) {
	userID := c.Param("user_id")
	stats, err := h.svc.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("get stats failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ActivateReferral POST /api/v1/referral/activate
func (h *Handler) ActivateReferral(c *gin.Context) {
	var req struct {
		RefCode  string `json:"ref_code"  binding:"required"`
		UserID   string `json:"user_id"   binding:"required"`
		Platform string `json:"platform"`
		DeviceID string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	resp, err := h.svc.ActivateReferral(c.Request.Context(), ports.ActivateReferralRequest{
		RefCode:  req.RefCode,
		UserID:   req.UserID,
		Platform: req.Platform,
		DeviceID: req.DeviceID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSelfReferral) {
			c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
			return
		}
		h.log.Error("activate referral failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"attributed":    resp.Attributed,
		"owner_user_id": resp.OwnerUserID,
		"reward_type":   resp.RewardType,
	})
}

// GetReferralStatus GET /api/v1/referral/status?ref_code=ABC123
func (h *Handler) GetReferralStatus(c *gin.Context) {
	refCode := c.Query("ref_code")
	if refCode == "" {
		c.JSON(http.StatusBadRequest, errorResponse("ref_code is required"))
		return
	}

	status, err := h.svc.GetReferralStatus(c.Request.Context(), refCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorResponse("link not found"))
			return
		}
		h.log.Error("get referral status failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, status)
}

func errorResponse(msg string) gin.H {
	return gin.H{"error": msg}
}
