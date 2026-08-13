package http

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	SDK        *SDKHandler
	Links      *LinksHandler
	Redirect   *RedirectHandler
	WellKnown  *WellKnownHandler
	Health     *HealthHandler
	Analytics  *AnalyticsHandler
	Webhooks   *WebhooksHandler
	Templates  *TemplatesHandler
	QR         *QRHandler
	CORSOrigin string
}

func NewRouter(h Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	if h.CORSOrigin == "*" || h.CORSOrigin == "" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = splitCommaList(h.CORSOrigin)
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(corsConfig))

	r.GET("/health", h.Health.Live)
	r.GET("/health/ready", h.Health.Ready)

	r.GET("/.well-known/apple-app-site-association", h.WellKnown.AppleAppSiteAssociation)
	r.GET("/.well-known/assetlinks.json", h.WellKnown.AssetLinks)

	api := r.Group("/api")
	{
		api.GET("/links", h.Links.List)
		api.POST("/links", h.Links.Create)
		api.GET("/links/:id", h.Links.Get)
		api.PUT("/links/:id", h.Links.Update)
		api.DELETE("/links/:id", h.Links.Delete)
		api.POST("/links/:id/duplicate", h.Links.Duplicate)
		api.GET("/links/:id/qr", h.QR.Generate)

		api.GET("/analytics/overview", h.Analytics.Overview)
		api.GET("/analytics/links/:linkId", h.Analytics.ForLink)

		sdk := api.Group("/sdk/v1")
		{
			sdk.POST("/install", h.SDK.Install)
			sdk.GET("/resolve/:shortCode", h.SDK.ResolveByShortCode)
			sdk.GET("/resolve/:templateSlug/:shortCode", h.SDK.ResolveByTemplateAndShortCode)
			sdk.POST("/event", h.SDK.Event)
			sdk.GET("/attribution/:fingerprint", h.SDK.Attribution)
			sdk.GET("/health", h.SDK.Health)
		}

		api.GET("/webhooks", h.Webhooks.List)
		api.POST("/webhooks", h.Webhooks.Create)
		api.GET("/webhooks/:id", h.Webhooks.Get)
		api.PUT("/webhooks/:id", h.Webhooks.Update)
		api.DELETE("/webhooks/:id", h.Webhooks.Delete)
		api.POST("/webhooks/:id/test", h.Webhooks.Test)

		api.GET("/templates", h.Templates.List)
		api.POST("/templates", h.Templates.Create)
		api.GET("/templates/:id", h.Templates.Get)
		api.PUT("/templates/:id", h.Templates.Update)
		api.DELETE("/templates/:id", h.Templates.Delete)
		api.PUT("/templates/:id/set-default", h.Templates.SetDefault)
	}

	// Catch-all short-link routes -- registered last, mirroring upstream.
	r.GET("/:templateSlug/:shortCode", h.Redirect.TemplateAndShortCode)
	r.GET("/:shortCode", h.Redirect.ShortCode)

	return r
}

func splitCommaList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
