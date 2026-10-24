package consumer

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log = newLogger()

func newLogger() *zap.SugaredLogger {
	cfg := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		OutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			TimeKey:      "time",
			LevelKey:     "level",
			CallerKey:    "caller",
			EncodeCaller: zapcore.FullCallerEncoder,
			EncodeLevel: func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString("[" + level.CapitalString() + "]")
			},
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02 15:04:05"))
			},
		},
	}
	l, err := cfg.Build()
	if err != nil {
		return zap.NewNop().Sugar()
	}
	return l.Sugar()
}
