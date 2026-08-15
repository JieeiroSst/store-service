package logger

import (
	"context"

	"github.com/JIeeiroSst/photo-service/pkg/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(lc fx.Lifecycle, cfg *config.Config) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	_ = level.UnmarshalText([]byte(cfg.App.LogLevel))

	var zcfg zap.Config
	if cfg.App.Env == "production" {
		zcfg = zap.NewProductionConfig()
	} else {
		zcfg = zap.NewDevelopmentConfig()
	}
	zcfg.Level = zap.NewAtomicLevelAt(level)

	l, err := zcfg.Build()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_ = l.Sync()
			return nil
		},
	})

	return l, nil
}

func FxLogger(l *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: l}
}
