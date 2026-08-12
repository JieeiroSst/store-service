package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    *config.Config
	Logger    *zap.Logger
	Router    http.Handler `name:"httpRouter"`
}

func NewHTTPServer(p Params) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", p.Config.Server.Port),
		Handler: p.Router,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p.Logger.Info("starting http server", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					p.Logger.Error("http server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("stopping http server")
			return srv.Shutdown(ctx)
		},
	})
}
