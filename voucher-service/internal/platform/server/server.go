package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Serve(lc fx.Lifecycle, srv *http.Server, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("http server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			log.Info("shutting down http server")
			return srv.Shutdown(shutdownCtx)
		},
	})
}
