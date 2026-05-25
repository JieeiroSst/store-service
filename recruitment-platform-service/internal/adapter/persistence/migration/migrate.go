package migration

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed *.sql
var migrationFiles embed.FS

// Run applies all pending migrations from the migrations directory.
func Run(ctx context.Context, db *sql.DB, logger *zap.Logger) error {
	goose.SetBaseFS(migrationFiles)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migration: set dialect: %w", err)
	}

	current, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("migration: get version: %w", err)
	}
	logger.Info("current schema version", zap.Int64("version", current))

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("migration: up: %w", err)
	}

	latest, _ := goose.GetDBVersion(db)
	logger.Info("migrations applied", zap.Int64("version", latest))
	return nil
}

func Rollback(ctx context.Context, db *sql.DB, steps int, logger *zap.Logger) error {
	goose.SetBaseFS(migrationFiles)
	_ = goose.SetDialect("postgres")
	for i := 0; i < steps; i++ {
		if err := goose.DownContext(ctx, db, "."); err != nil {
			return fmt.Errorf("migration: down step %d: %w", i+1, err)
		}
	}
	v, _ := goose.GetDBVersion(db)
	logger.Info("rollback complete", zap.Int64("version", v))
	return nil
}
