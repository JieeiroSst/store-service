package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func New(cfg *config.Config) (*gorm.DB, error) {
	m := cfg.Mysql
	dsn := func(db string) string {
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			m.User, m.Password, m.Host, m.Port, db)
	}

	if err := ensureDatabase(dsn(""), m.DBName); err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dsn(m.DBName)), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(model.Models()...); err != nil {
		return nil, err
	}
	return db, nil
}

func ensureDatabase(serverDSN, name string) error {
	conn, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec("CREATE DATABASE IF NOT EXISTS `" + name + "` CHARACTER SET utf8mb4")
	return err
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
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)
