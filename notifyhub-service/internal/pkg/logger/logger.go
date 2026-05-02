package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(mode string) (*zap.Logger, error) {
	var cfg zap.Config
	if mode == "development" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return cfg.Build()
}

func Must(mode string) *zap.Logger {
	l, err := New(mode)
	if err != nil {
		// Fallback to basic logger
		l, _ = zap.NewProduction()
		l.Error("failed to build logger, using default", zap.Error(err))
	}
	return l
}

func WithRequestID(l *zap.Logger, id string) *zap.Logger {
	return l.With(zap.String("request_id", id))
}

func WithJobID(l *zap.Logger, id string) *zap.Logger {
	return l.With(zap.String("job_id", id))
}

func Hostname() string {
	h, _ := os.Hostname()
	return h
}
