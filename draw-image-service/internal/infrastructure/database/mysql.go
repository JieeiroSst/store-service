package database

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/draw-image-service/config"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Module = fx.Options(
	fx.Provide(NewMysql),
)

func NewMysql(lc fx.Lifecycle, cfg *config.Config) (*gorm.DB, error) {
	m := cfg.Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		m.MysqlUser, m.MysqlPassword, m.MysqlHost, m.MysqlPort, m.MysqlDbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
	return db, nil
}
