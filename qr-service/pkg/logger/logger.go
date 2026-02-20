package logger

import (
	"github.com/JIeeiroSst/qr-service/internal/infrastructure/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapCfg zap.Config

	if cfg.App.Env == "production" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	return zapCfg.Build()
}
