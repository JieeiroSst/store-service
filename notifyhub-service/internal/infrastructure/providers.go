package infrastructure

import (
	"github.com/JIeeiroSst/notifyhub-service/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/application"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newConfig() *config.Config {
	return config.Load(config.Path())
}

func newLogger(cfg *config.Config) (*zap.Logger, error) {
	var zc zap.Config
	if cfg.Server.Mode == "development" {
		zc = zap.NewDevelopmentConfig()
		zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zc = zap.NewProductionConfig()
	}
	zc.OutputPaths = []string{"stdout"}
	zc.ErrorOutputPaths = []string{"stderr"}
	log, err := zc.Build()
	if err != nil {
		return nil, err
	}
	return log.With(zap.String("service", "notifyhub-service")), nil
}

func FxLogger(log *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zapcore.WarnLevel))}
}

func newSettings(cfg *config.Config) application.Settings {
	s := application.DefaultSettings()
	s.RetryMax = cfg.Worker.RetryMax
	s.RetryDelay = cfg.Worker.RetryDelayDuration()
	return s
}
