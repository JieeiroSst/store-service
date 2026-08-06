package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/JIeeiroSst/cdn-service/config"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

func NewDatabase(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.PostgresqlHost,
		cfg.Postgres.PostgresqlPort,
		cfg.Postgres.PostgresqlUser,
		cfg.Postgres.PostgresqlPassword,
		cfg.Postgres.PostgresqlDbname,
		sslMode(cfg.Postgres.PostgresqlSSLMode),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func sslMode(enabled bool) string {
	if enabled {
		return "require"
	}
	return "disable"
}

func registerLifecycle(lc fx.Lifecycle, db *sql.DB) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return db.Close()
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
	fx.Invoke(registerLifecycle),
)
