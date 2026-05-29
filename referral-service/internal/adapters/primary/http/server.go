package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	appconfig "github.com/referral/service/internal/config"
	"github.com/referral/service/pkg/logger"
)

var ServerModule = fx.Options(
	fx.Provide(NewServer),
	fx.Invoke(RegisterRoutes),
)

func NewServer(cfg *appconfig.Config, log *zap.Logger) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(logger.GinMiddleware(log))
	r.Use(gin.Recovery())

	return r
}

func RegisterRoutes(
	lc fx.Lifecycle,
	r *gin.Engine,
	h *Handler,
	cfg *appconfig.Config,
	log *zap.Logger,
) {
	h.Register(r)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info("HTTP server starting",
				zap.String("addr", addr),
				zap.String("env", cfg.App.Env),
			)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal("HTTP server crashed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("HTTP server shutting down gracefully")
			return srv.Shutdown(ctx)
		},
	})
}
