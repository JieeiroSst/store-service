package database

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/JIeeiroSst/post-service/config"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	instance *gorm.DB
	once     sync.Once
	openErr  error
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Dbname)
		instance, openErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	})
	if openErr != nil {
		return nil, openErr
	}

	if err := applySchema(instance, "database.sql"); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return instance, nil
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
