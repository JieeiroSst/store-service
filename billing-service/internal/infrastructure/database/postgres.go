package database

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/billing-service/config"
	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	sslMode := "disable"
	if cfg.Postgres.PostgresqlSSLMode {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.PostgresqlHost,
		cfg.Postgres.PostgresqlPort,
		cfg.Postgres.PostgresqlUser,
		cfg.Postgres.PostgresqlPassword,
		cfg.Postgres.PostgresqlDbname,
		sslMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Address{},
		&model.Customer{},
		&model.Plan{},
		&model.Subscription{},
		&model.Invoice{},
		&model.Transaction{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

func registerLifecycle(lc fx.Lifecycle, db *gorm.DB) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
	fx.Invoke(registerLifecycle),
)
