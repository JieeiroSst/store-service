package secondary

import (
	"context"
	"log/slog"

	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/secondary/postgres"
	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/secondary/uploadservice"
	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewRepositories),
)

type Repositories struct {
	fx.Out

	Users    port.UserRepository
	Roster   port.RosterRepository
	Offline  port.OfflineRepository
	Health   port.HealthChecker
	External port.ExternalAuthenticator
	Files    port.FileStorage
}

type noopHealth struct{}

func (noopHealth) Ping(context.Context) error { return nil }

func NewRepositories(lc fx.Lifecycle, cfg *config.Config) (Repositories, error) {
	out := Repositories{
		Users:   memory.NewUserStore(),
		Roster:  memory.NewRosterStore(),
		Offline: memory.NewOfflineStore(),
		Health:  noopHealth{},
	}
	if cfg.Modules.HTTPUpload.Enabled {
		out.Files = uploadservice.New(cfg.UploadService)
		slog.Info("storing uploaded files in upload_service", "url", cfg.UploadService.URL)
	}
	if cfg.ExternalAuth() {
		out.External = userservice.New(cfg.UserService)
		slog.Info("authenticating users with user_service", "url", cfg.UserService.URL, "jid_field", cfg.UserService.JIDField)
	}
	if !cfg.UseSQL() {
		slog.Warn("using in-memory storage (default_db: mnesia); data is lost on restart")
		return out, nil
	}
	db, err := postgres.Open(cfg)
	if err != nil {
		return out, err
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return db.Migrate(ctx, cfg.AuthMethod == "sql") },
		OnStop:  func(context.Context) error { return db.Close() },
	})
	out.Health = db
	if cfg.AuthMethod == "sql" {
		out.Users = postgres.NewUserStore(db)
	}
	if cfg.DefaultDB == "sql" {
		out.Roster = postgres.NewRosterStore(db)
		out.Offline = postgres.NewOfflineStore(db)
	}
	slog.Info("using postgres storage", "server", cfg.SQL.Server, "database", cfg.SQL.Database,
		"auth", cfg.AuthMethod, "default_db", cfg.DefaultDB)
	return out, nil
}
