package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/shortlink-service/internal/config"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	migrationsrc "github.com/JIeeiroSst/shortlink-service/migrations"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/fx"
	"go.uber.org/zap"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var Module = fx.Module("db",
	fx.Provide(NewGormDB),
	fx.Provide(func(db *gorm.DB) ports.LinkRepository { return NewLinkRepo(db) }),
	fx.Provide(func(db *gorm.DB) ports.ClickEventRepository { return NewClickRepo(db) }),
	fx.Provide(func(db *gorm.DB) ports.FingerprintRepository { return NewFingerprintRepo(db) }),
	fx.Provide(func(db *gorm.DB) ports.InstallEventRepository { return NewInstallRepo(db) }),
	fx.Provide(func(db *gorm.DB) ports.InAppEventRepository { return NewInAppEventRepo(db) }),
	fx.Provide(func(db *gorm.DB) ports.WebhookRepository { return NewWebhookRepo(db) }),
	fx.Provide(func(db *gorm.DB) ports.TemplateRepository { return NewTemplateRepo(db) }),
	fx.Provide(func(db *gorm.DB) ports.DBPing { return NewDBPing(db) }),
)

func NewGormDB(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(gormpostgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := sqlDB.PingContext(ctx); err != nil {
				return fmt.Errorf("ping database: %w", err)
			}
			if err := runMigrations(cfg.DatabaseURL, log); err != nil {
				return fmt.Errorf("run migrations: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})

	return db, nil
}

func runMigrations(databaseURL string, log *zap.Logger) error {
	sourceDriver, err := iofs.New(migrationsrc.FS, ".")
	if err != nil {
		return err
	}

	migrateURL := "postgres" + strings.TrimPrefix(databaseURL, "postgresql")

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, migrateURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	log.Info("database migrations applied")
	return nil
}
