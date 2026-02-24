package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func Init(env string) {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	zapLogger, err := config.Build()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	Logger = zapLogger.Sugar()
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
