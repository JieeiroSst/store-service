package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/fx"

	"github.com/JIeeiroSst/bonuslink-service/internal/config"
	"github.com/JIeeiroSst/bonuslink-service/internal/core/ports"
)

var Module = fx.Options(
	fx.Provide(
		NewClient,
		fx.Annotate(NewRewardRepo, fx.As(new(ports.RewardRepository))),
	),
	fx.Invoke(registerLifecycle),
)

func NewClient(cfg *config.Config) (*sqlx.DB, error) {
	p := cfg.Postgres
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Database, p.SSLMode)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	db.SetMaxOpenConns(p.MaxOpenConns)
	db.SetMaxIdleConns(p.MaxIdleConns)
	db.SetConnMaxLifetime(p.ConnMaxLifetime)
	return db, nil
}

func registerLifecycle(lc fx.Lifecycle, db *sqlx.DB) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error { return db.Close() },
	})
}
