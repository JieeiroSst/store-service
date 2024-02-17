package logger

import (
	"go.uber.org/zap"
)

func Logger() *zap.Logger {
	logger, _ := zap.NewProductionConfig().Build()

	return logger
}