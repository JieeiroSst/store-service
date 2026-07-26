package database

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Mysql.MysqlUser,
		cfg.Mysql.MysqlPassword,
		cfg.Mysql.MysqlHost,
		cfg.Mysql.MysqlPort,
		cfg.Mysql.MysqlDbname,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Notification{},
		&model.UserDevice{},
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
