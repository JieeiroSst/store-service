package database

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/delivery-service/config"
	"github.com/JIeeiroSst/delivery-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Postgres.PostgresqlHost, cfg.Postgres.PostgresqlUser, cfg.Postgres.PostgresqlPassword,
		cfg.Postgres.PostgresqlDbname, cfg.Postgres.PostgresqlPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.Delivery{}); err != nil {
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
