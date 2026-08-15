package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/photo-service/pkg/config"
)

func RegisterServer(lc fx.Lifecycle, router *gin.Engine, cfg *config.Config, log *zap.Logger) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return fmt.Errorf("listen on %s: %w", srv.Addr, err)
			}
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Error("http server stopped unexpectedly", zap.Error(err))
				}
			}()
			log.Info("http server started", zap.String("addr", srv.Addr))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.HTTP.ShutdownTimeout))
			defer cancel()
			log.Info("shutting down http server")
			return srv.Shutdown(shutdownCtx)
		},
	})
}
