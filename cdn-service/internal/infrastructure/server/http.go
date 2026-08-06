package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/cdn-service/config"
	httpadapter "github.com/JIeeiroSst/cdn-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type HTTPParams struct {
	fx.In

	LC      fx.Lifecycle
	Handler *httpadapter.Handler
	Config  *config.Config
}

// NewHTTPServer serves static/uploaded file content over plain Gin routes,
// separate from the grpc-gateway REST API.
func NewHTTPServer(p HTTPParams) {
	engine := httpadapter.NewRouter(p.Handler)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", p.Config.Server.PortGinServer),
		Handler: engine,
	}

	p.LC.Append(fx.Hook{
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
