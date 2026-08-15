package main

import (
	"github.com/JIeeiroSst/voucher-service/internal/app"
	"github.com/JIeeiroSst/voucher-service/internal/platform/logger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		app.Module,
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return logger.FxLogger(log)
		}),
	).Run()
}
