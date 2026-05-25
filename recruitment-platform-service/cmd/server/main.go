package main

import (
	"context"
	"os"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/module"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/config.yaml"
	}

	app := fx.New(
		fx.Provide(func() (*module.Config, error) {
			return module.LoadConfig(cfgPath)
		}),

		module.LoggerModule,
		module.DatabaseModule,
		module.InfraModule,
		module.TemporalModule,
		module.ServiceModule,
		module.HTTPModule,

		fx.Invoke(func(lc fx.Lifecycle, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("🚀 Recruitment Platform started")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("👋 Recruitment Platform stopped gracefully")
					return nil
				},
			})
		}),
	)

	app.Run()
}
