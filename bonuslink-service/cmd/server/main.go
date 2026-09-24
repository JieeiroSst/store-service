package main

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/bonuslink-service/internal/adapters/primary/http"
	"github.com/JIeeiroSst/bonuslink-service/internal/adapters/primary/queue"
	"github.com/JIeeiroSst/bonuslink-service/internal/adapters/secondary/postgres"
	"github.com/JIeeiroSst/bonuslink-service/internal/config"
	"github.com/JIeeiroSst/bonuslink-service/internal/core/services"
	"github.com/JIeeiroSst/bonuslink-service/pkg/logger"
)

func newLoggerConfig(cfg *config.Config) *logger.Config {
	return &logger.Config{
		AppEnv:     cfg.App.Env,
		AppName:    cfg.App.Name,
		AppVersion: cfg.App.Version,
		Level:      cfg.Logger.Level,
		FilePath:   cfg.Logger.FilePath,
		MaxSizeMB:  cfg.Logger.MaxSizeMB,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAgeDays: cfg.Logger.MaxAgeDays,
	}
}

func main() {
	fx.New(
		config.Module,
		fx.Provide(newLoggerConfig),
		logger.Module,
		postgres.Module,
		services.Module,
		http.Module,
		http.ServerModule,
		queue.Module,
	).Run()
}
