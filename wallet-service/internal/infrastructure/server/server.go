package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/config"
)

func New(handler http.Handler, cfg *config.Config) *http.Server {
	return &http.Server{
		Addr:              cfg.Server.Port,
		Handler:           handler,
		ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func Register(lc fx.Lifecycle, shutdowner fx.Shutdowner, srv *http.Server, cfg *config.Config, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("wallet-service listening", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("http server stopped", zap.Error(err))
					_ = shutdowner.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("wallet-service shutting down")
			ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
			defer cancel()
			return srv.Shutdown(ctx)
		},
	})
}
