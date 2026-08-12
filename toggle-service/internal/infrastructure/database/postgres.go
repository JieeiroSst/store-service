package database

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    *config.Config
}

func NewDatabase(p Params) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(p.Config.Postgres.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})

	return db, nil
}

var Module = fx.Options(fx.Provide(NewDatabase))
