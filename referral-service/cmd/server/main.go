package main

import (
	"go.uber.org/fx"

	"github.com/referral/service/internal/adapters/primary/http"
	"github.com/referral/service/internal/adapters/secondary/cache"
	"github.com/referral/service/internal/adapters/secondary/mysql"
	"github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/services"
	"github.com/referral/service/pkg/logger"
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
	app := fx.New(
		config.Module,
		fx.Provide(newLoggerConfig),
		logger.Module,
		mysql.Module,
		cache.Module,
		services.Module,
		http.Module,
		http.ServerModule,
	)

	app.Run()
}
