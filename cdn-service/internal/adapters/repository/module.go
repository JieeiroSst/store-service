package repository

import (
	"github.com/JIeeiroSst/cdn-service/internal/config"
	"github.com/JIeeiroSst/cdn-service/internal/ports"
	"go.uber.org/fx"
)

func runMigrations(cfg *config.Config) error {
	return RunMigrations(cfg)
}

var Module = fx.Options(
	fx.Provide(NewPostgresDB),
	fx.Invoke(registerLifecycle),
	fx.Invoke(runMigrations),
	fx.Provide(NewFileRepository),
	fx.Provide(func(r *FileRepository) ports.FileRepository { return r }),
)
