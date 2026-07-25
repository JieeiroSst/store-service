package database

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/address-country-service/config"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.Mysql.MysqlUser,
		cfg.Mysql.MysqlPassword,
		cfg.Mysql.MysqlHost,
		cfg.Mysql.MysqlPort,
		cfg.Mysql.MysqlDbname,
	)

	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
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
