package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Module = fx.Options(
	fx.Provide(New),
)

func New(lc fx.Lifecycle, cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Postgres.DSN()), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.Postgres.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Postgres.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error { return sqlDB.Close() },
	})
	return db, nil
}
