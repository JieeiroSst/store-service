package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/qr-service/internal/infrastructure/config"
	"github.com/JIeeiroSst/qr-service/internal/module"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		module.AllModules,
		fx.Invoke(startServer),
	)
	app.Run()
}

func startServer(
	lc fx.Lifecycle,
	router *gin.Engine,
	cfg *config.Config,
	logger *zap.Logger,
) {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting QR service",
				zap.String("port", cfg.Server.Port),
				zap.String("base_url", cfg.App.BaseURL),
				zap.String("env", cfg.App.Env),
			)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down server gracefully...")
			return server.Shutdown(ctx)
		},
	})
}
