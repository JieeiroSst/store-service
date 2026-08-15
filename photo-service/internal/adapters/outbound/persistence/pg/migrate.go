package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/photo-service/migrations"
	"github.com/JIeeiroSst/photo-service/pkg/config"
)

func RegisterMigrations(lc fx.Lifecycle, _ *pgxpool.Pool, cfg *config.Config, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			db, err := sql.Open("pgx", cfg.Postgres.DSN())
			if err != nil {
				return fmt.Errorf("open migration db connection: %w", err)
			}
			defer db.Close()

			driver, err := postgres.WithInstance(db, &postgres.Config{})
			if err != nil {
				return fmt.Errorf("init postgres migration driver: %w", err)
			}

			src, err := iofs.New(migrations.FS, ".")
			if err != nil {
				return fmt.Errorf("load embedded migrations: %w", err)
			}

			m, err := migrate.NewWithInstance("iofs", src, "pgx", driver)
			if err != nil {
				return fmt.Errorf("init migrator: %w", err)
			}

			if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("run migrations: %w", err)
			}

			log.Info("database migrations applied")
			return nil
		},
	})
}
