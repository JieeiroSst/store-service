// Command server is the LinkForty Core (Go port) entrypoint: hexagonal
// architecture wired together with uber-go/fx. Each concern (config, db,
// cache, geoip, webhook sender, qrcode, services, router) is its own
// fx.Module; see internal/*/module.go.
package main

import (
	"github.com/JIeeiroSst/shortlink-service/internal/adapters/cache"
	"github.com/JIeeiroSst/shortlink-service/internal/adapters/geoip"
	httpadapter "github.com/JIeeiroSst/shortlink-service/internal/adapters/http"
	"github.com/JIeeiroSst/shortlink-service/internal/adapters/qrcode"
	"github.com/JIeeiroSst/shortlink-service/internal/adapters/repo"
	"github.com/JIeeiroSst/shortlink-service/internal/adapters/webhook"
	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/JIeeiroSst/shortlink-service/internal/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(newLogger),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),

		config.Module,
		repo.Module,
		cache.Module,
		geoip.Module,
		webhook.Module,
		qrcode.Module,
		app.Module,
		httpadapter.Module,
	).Run()
}

func newLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}
