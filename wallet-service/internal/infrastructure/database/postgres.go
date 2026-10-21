package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/config"
	"github.com/JIeeiroSst/wallet-service/migrations"
)

func NewPostgres(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (*sqlx.DB, error) {
	dsn, err := cfg.Database.ConnString()
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if cfg.Database.AutoMigrate {
		if err := Migrate(ctx, db, migrations.Schema, log); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	if cfg.Database.SeedDevData {
		if err := Seed(ctx, db, migrations.Seed, "seed", log); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	lc.Append(fx.StopHook(db.Close))
	return db, nil
}
