package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_init.sql
var initSQL string

const migrateLockID = 7302026

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, cfg.PostgresURL(cfg.Postgres.Database))
}

func ensureDatabase(ctx context.Context, cfg *config.Config) error {
	conn, err := pgx.Connect(ctx, cfg.PostgresURL("postgres"))
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer conn.Close(ctx)

	var exists bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, cfg.Postgres.Database).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{cfg.Postgres.Database}.Sanitize()); err != nil {
		var again bool
		if qerr := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, cfg.Postgres.Database).Scan(&again); qerr == nil && again {
			return nil
		}
		return err
	}
	return nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrateLockID); err != nil {
		return err
	}
	defer conn.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, migrateLockID)

	_, err = conn.Conn().PgConn().Exec(ctx, initSQL).ReadAll()
	return err
}
