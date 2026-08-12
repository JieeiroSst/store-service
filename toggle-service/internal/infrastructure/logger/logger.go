package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.Server.Env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

var Module = fx.Options(fx.Provide(NewLogger))
