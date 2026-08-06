package database

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/bookStore-service/config"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.PostgresqlHost,
		cfg.Postgres.PostgresqlPort,
		cfg.Postgres.PostgresqlUser,
		cfg.Postgres.PostgresqlPassword,
		cfg.Postgres.PostgresqlDbname,
		sslMode(cfg.Postgres.PostgresqlSSLMode),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Parent tables first so the foreign keys on book_store_book resolve.
	if err := db.AutoMigrate(
		&model.Author{},
		&model.Publisher{},
		&model.Category{},
		&model.Book{},
	); err != nil {
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
