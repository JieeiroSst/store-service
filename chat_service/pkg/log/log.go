package log

import (
	"time"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	timeFormat = "2006-01-02 15:04:05"
)

func configZap() *zap.SugaredLogger {
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
			EncodeLevel:  customLevelEncoder,
			EncodeTime:   syslogTimeEncoder,
		},
	}

	logger, _ := cfg.Build()
	return logger.Sugar()
}

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(timeFormat))
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

var sugarLogger = configZap()

func Info(msg interface{}) {
	msgJson, _ := json.Marshal(msg)
	sugarLogger.Infof("log is :%s %v", time.Now().Format(timeFormat), msgJson)
}

func Debug(msg interface{}) {
	msgJson, _ := json.Marshal(msg)
	sugarLogger.Debugf("log is :%s %v", time.Now().Format(timeFormat), msgJson)
}

func Warning(msg interface{}) {
	msgJson, _ := json.Marshal(msg)
	sugarLogger.Warnf("log is :%s %v", time.Now().Format(timeFormat), msgJson)
}

func Error(msg interface{}) {
	msgJson, _ := json.Marshal(msg)
	sugarLogger.Errorf("log is :%s %v", time.Now().Format(timeFormat), msgJson)
}
