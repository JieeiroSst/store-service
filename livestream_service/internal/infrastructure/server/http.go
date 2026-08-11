package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/livestream-service/config"
	httpadapter "github.com/JIeeiroSst/livestream-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func serve(lc fx.Lifecycle, cfg *config.Config, engine http.Handler) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.PortHttpServer),
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.WithContext(ctx).Error("http server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}

// NewNodeHTTPServer serves only the SRS http_hooks contract - see
// httpadapter.NewNodeRouter.
type NodeHTTPParams struct {
	fx.In

	LC     fx.Lifecycle
	SRS    *httpadapter.SRSWebhookHandler
	Config *config.Config
}

func NewNodeHTTPServer(p NodeHTTPParams) {
	serve(p.LC, p.Config, httpadapter.NewNodeRouter(p.SRS))
}

// NewEdgeHTTPServer serves rooms/ingest/viewers/chat - see
// httpadapter.NewEdgeRouter.
type EdgeHTTPParams struct {
	fx.In

	LC      fx.Lifecycle
	Handler *httpadapter.Handler
	WS      *httpadapter.WSHandler
	Config  *config.Config
}

func NewEdgeHTTPServer(p EdgeHTTPParams) {
	serve(p.LC, p.Config, httpadapter.NewEdgeRouter(p.Handler, p.WS))
}

// NewHTTPServer serves every route on one port - only used by the
// monolithic dev/docker-compose entrypoint (cmd/main.go).
type HTTPParams struct {
	fx.In

	LC      fx.Lifecycle
	Handler *httpadapter.Handler
	SRS     *httpadapter.SRSWebhookHandler
	WS      *httpadapter.WSHandler
	Config  *config.Config
}

func NewHTTPServer(p HTTPParams) {
	serve(p.LC, p.Config, httpadapter.NewAllInOneRouter(p.Handler, p.SRS, p.WS))
}
