package infrastructure

import (
	"log/slog"
	"os"
	"strings"

	httpapi "github.com/JIeeiroSst/ejabberd-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/primary/xmpp"
	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/secondary"
	"github.com/JIeeiroSst/ejabberd-service/internal/application"
	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/infrastructure/server"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var Module = fx.Options(
	fx.Provide(config.Load),
	CoreModule,
)

var CoreModule = fx.Options(
	fx.Provide(
		func(cfg *config.Config) *domain.ACLMatcher { return cfg.Matcher() },
		NewTLSConfig,
	),
	fx.Invoke(setupLogging),

	secondary.Module,

	application.Module,

	xmpp.Module,
	httpapi.Module,

	fx.Invoke(server.NewHTTP),
)

func Logger() fx.Option {
	return fx.WithLogger(func() fxevent.Logger {
		l := &fxevent.SlogLogger{Logger: slog.Default().With("component", "fx")}
		l.UseLogLevel(slog.LevelDebug)
		return l
	})
}

func setupLogging(cfg *config.Config) {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warning", "warn":
		level = slog.LevelWarn
	case "error", "critical", "alert", "emergency":
		level = slog.LevelError
	case "none":
		level = slog.LevelError + 4
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
	slog.Info("starting ejabberd", "version", config.Version, "hosts", cfg.Hosts)
	if len(cfg.Unsupported) > 0 {
		slog.Warn("config options not supported by this port are ignored", "options", cfg.Unsupported)
	}
}
