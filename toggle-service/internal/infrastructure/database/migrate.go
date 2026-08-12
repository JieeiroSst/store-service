package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
)

const migrationsSourceURL = "file://migrations"

func RunMigrations(cfg *config.Config, log *zap.Logger) error {
	m, err := migrate.New(migrationsSourceURL, cfg.Postgres.URL())
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Info("database migrations applied")
	return nil
}
