package database

import (
	"fmt"
	"os"
	"strings"

	"github.com/JIeeiroSst/payment-service/config"
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
	if err := applySchema(db, "database.sql"); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return db, nil
}

func applySchema(db *gorm.DB, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, stmt := range strings.Split(string(raw), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("exec statement %q: %w", stmt, err)
		}
	}
	return nil
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
)
