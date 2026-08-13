package http

import (
	"net/http"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type RedirectHandler struct {
	redirect           *app.RedirectService
	trustProxy         domain.TrustProxy
	trustEdgeBotHeader bool
	abuseReportURL     string
}

func NewRedirectHandler(redirect *app.RedirectService, trustProxy domain.TrustProxy, trustEdgeBotHeader bool, abuseReportURL string) *RedirectHandler {
	return &RedirectHandler{redirect, trustProxy, trustEdgeBotHeader, abuseReportURL}
}

func (h *RedirectHandler) handle(c *gin.Context, shortCode, templateSlug string) {
	in := app.RedirectInput{
		ShortCode: shortCode, TemplateSlug: templateSlug,
		IP: ClientIP(c, h.trustProxy), UserAgent: c.Request.UserAgent(),
		Referrer: c.GetHeader("Referer"), AcceptLanguage: c.GetHeader("Accept-Language"),
		Method:             c.Request.Method,
		TrustEdgeBotHeader: h.trustEdgeBotHeader, EdgeBotHeaderValue: c.GetHeader("X-LF-Bot"),
		UTMSource: c.Query("utm_source"), UTMMedium: c.Query("utm_medium"), UTMCampaign: c.Query("utm_campaign"),
		FPTimezone: c.Query("fp_tz"), FPLanguage: c.Query("fp_lang"),
		FPScreenWidth: c.Query("fp_sw"), FPScreenHeight: c.Query("fp_sh"),
		AbuseReportURL: h.abuseReportURL,
	}

	outcome, err := h.redirect.Resolve(c.Request.Context(), in)
	if err != nil {
		respondInternalError(c, "Failed to resolve redirect", err)
		return
	}

	switch outcome.Kind {
	case app.OutcomeNotFound:
		respondNotFound(c, "Link not found")
	case app.OutcomeNoDestination:
		c.JSON(http.StatusNotFound, gin.H{"error": "No destination URL configured for this link"})
	case app.OutcomeWarnHTML:
		c.Header("X-Robots-Tag", "noindex, nofollow")
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(outcome.HTML))
	case app.OutcomeInterstitialHTML:
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(outcome.HTML))
	case app.OutcomeRedirect:
		c.Redirect(http.StatusFound, outcome.RedirectURL)
	}
}

// TemplateAndShortCode handles GET /:templateSlug/:shortCode.
func (h *RedirectHandler) TemplateAndShortCode(c *gin.Context) {
	h.handle(c, c.Param("shortCode"), c.Param("templateSlug"))
}

// ShortCode handles GET /:shortCode.
func (h *RedirectHandler) ShortCode(c *gin.Context) {
	h.handle(c, c.Param("shortCode"), "")
}
