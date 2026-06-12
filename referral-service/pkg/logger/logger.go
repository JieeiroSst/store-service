package logger

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Module = fx.Options(
	fx.Provide(
		New,
		NewSugared,
	),
	fx.Invoke(registerShutdownHook),
)

func New(cfg *Config) (*zap.Logger, error) {
	level := parseLevel(cfg.Level)

	var cores []zapcore.Core
	var opts []zap.Option

	if cfg.AppEnv != "production" {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			level,
		))
		opts = append(opts,
			zap.Development(),
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
		)
	} else {
		rawCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(jsonEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			level,
		)

		cores = append(cores, zapcore.NewSamplerWithOptions(
			rawCore,
			time.Second,
			100,
			100,
			zapcore.SamplerHook(func(e zapcore.Entry, dec zapcore.SamplingDecision) {
				if e.Level >= zapcore.ErrorLevel {
					dec = zapcore.LogSampled
				}
			}),
		))
	}

	if cfg.FilePath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0o755); err == nil {
			roller := &lumberjack.Logger{
				Filename:   cfg.FilePath,
				MaxSize:    cfg.MaxSizeMB,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAgeDays,
				Compress:   true,
			}
			cores = append(cores, zapcore.NewCore(
				zapcore.NewJSONEncoder(jsonEncoderConfig()),
				zapcore.AddSync(roller),
				level,
			))
		}
	}

	log := zap.New(zapcore.NewTee(cores...), opts...).With(
		zap.String("service", cfg.AppName),
		zap.String("env", cfg.AppEnv),
		zap.String("version", cfg.AppVersion),
	)

	log.Info("logger initialised",
		zap.String("level", cfg.Level),
		zap.String("env", cfg.AppEnv),
		zap.Bool("file_sink", cfg.FilePath != ""),
	)

	return log, nil
}

func NewSugared(log *zap.Logger) *zap.SugaredLogger {
	return log.Sugar()
}

func registerShutdownHook(lc fx.Lifecycle, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			_ = log.Sync()
			return nil
		},
	})
}

func consoleEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys
		TimeKey:       "T",
		LevelKey:      "L",
		NameKey:       "N",
		CallerKey:     "C",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "M",
		StacktraceKey: "S",

		// Formatting
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000Z07:00"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

func jsonEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",

		// Formatting
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // "info", "error"
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // "2024-01-15T10:05:32.123+0700"
		EncodeDuration: zapcore.MillisDurationEncoder, // 1.234 (numeric ms)
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

func parseLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}
