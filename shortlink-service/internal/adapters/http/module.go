package http

import (
	"context"
	"net"
	"net/http"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/JIeeiroSst/shortlink-service/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("router",
	fx.Provide(
		func(s *app.SDKService, cfg config.Config) *SDKHandler {
			return NewSDKHandler(s, cfg.TrustProxy, cfg.TrustEdgeBotHeader)
		},
		func(s *app.LinkService) *LinksHandler { return NewLinksHandler(s) },
		func(s *app.RedirectService, cfg config.Config) *RedirectHandler {
			return NewRedirectHandler(s, cfg.TrustProxy, cfg.TrustEdgeBotHeader, cfg.AbuseReportURL)
		},
		func(cfg config.Config) *WellKnownHandler {
			return NewWellKnownHandler(cfg.IOSTeamID, cfg.IOSBundleID, cfg.AndroidPackageName, cfg.AndroidSHA256Fingerprints)
		},
		func(s *app.HealthService) *HealthHandler { return NewHealthHandler(s) },
		func(s *app.AnalyticsService) *AnalyticsHandler { return NewAnalyticsHandler(s) },
		func(s *app.WebhookService) *WebhooksHandler { return NewWebhooksHandler(s) },
		func(s *app.TemplateService) *TemplatesHandler { return NewTemplatesHandler(s) },
		func(s *app.QRService, cfg config.Config) *QRHandler { return NewQRHandler(s, cfg.ShortlinkDomain) },
		func(
			sdk *SDKHandler, links *LinksHandler, redirect *RedirectHandler, wellKnown *WellKnownHandler,
			health *HealthHandler, analytics *AnalyticsHandler, webhooks *WebhooksHandler,
			templates *TemplatesHandler, qr *QRHandler, cfg config.Config,
		) Handlers {
			return Handlers{
				SDK: sdk, Links: links, Redirect: redirect, WellKnown: wellKnown, Health: health,
				Analytics: analytics, Webhooks: webhooks, Templates: templates, QR: qr,
				CORSOrigin: cfg.CORSOrigin,
			}
		},
		NewRouter,
	),
	fx.Invoke(registerHTTPServer),
)

func registerHTTPServer(lc fx.Lifecycle, engine *gin.Engine, cfg config.Config, log *zap.Logger) {
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: engine}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("LinkForty server listening", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Error("http server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
