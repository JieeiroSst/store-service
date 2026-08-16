package database

import (
	"context"
	"net/url"

	"github.com/JIeeiroSst/admanagement-service/config"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func buildDSN(cfg *config.Config) string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Postgres.User, cfg.Postgres.Password),
		Host:   cfg.Postgres.Host + ":" + cfg.Postgres.Port,
		Path:   "/" + cfg.Postgres.DBName,
	}
	q := u.Query()
	q.Set("sslmode", cfg.Postgres.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := buildDSN(cfg)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.AdCampaign{},
		&model.AdCategory{},
		&model.AdPosition{},
		&model.Ad{},
		&model.AdPositionMapping{},
		&model.AdImpression{},
		&model.AdClick{},
		&model.AdTargetingRule{},
		&model.AdPerformanceSummary{},
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
