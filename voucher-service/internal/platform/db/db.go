package db

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func DSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword,
		cfg.PostgresDBName, cfg.PostgresSSLMode)
}

func NewGormDB(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	gdb, err := gorm.Open(postgres.Open(DSN(cfg)), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return sqlDB.PingContext(ctx)
		},
		OnStop: func(ctx context.Context) error {
			log.Info("closing postgres connection pool")
			return sqlDB.Close()
		},
	})

	return gdb, nil
}
