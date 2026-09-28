package database

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/notifyhub-service/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.Database.DataSourceName()), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		TranslateError:         true,
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetimeDuration())
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxLifetimeDuration() / 2)

	if err := db.AutoMigrate(
		&model.Channel{},
		&model.DataSource{},
		&model.Template{},
		&model.NotifyJob{},
		&model.NotifyHistory{},
	); err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}
	log.Info("mysql connected",
		zap.Int("max_open", cfg.Database.MaxOpenConns),
		zap.Int("max_idle", cfg.Database.MaxIdleConns),
	)

	lc.Append(fx.StopHook(func(context.Context) error { return sqlDB.Close() }))
	return db, nil
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
)
