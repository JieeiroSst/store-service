package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/bot-service/config"
	httpadapter "github.com/JIeeiroSst/bot-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type HTTPParams struct {
	fx.In

	LC       fx.Lifecycle
	Handler  *httpadapter.Handler
	Facebook *httpadapter.FacebookHandler
	Content  *httpadapter.ContentHandler
	Config   *config.Config
}

func NewHTTPServer(p HTTPParams) {
	engine := httpadapter.NewRouter(p.Handler, p.Facebook, p.Content, p.Config)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", p.Config.Server.PortHttpServer),
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
