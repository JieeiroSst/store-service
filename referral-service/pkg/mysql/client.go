package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	appconfig "github.com/referral/service/internal/config"
)

func NewClient(config *appconfig.Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&charset=utf8mb4",
		config.MySQL.User,
		config.MySQL.Password,
		config.MySQL.Host,
		config.MySQL.Port,
		config.MySQL.Database,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}

	db.SetMaxOpenConns(config.MySQL.MaxOpenConns)
	db.SetMaxIdleConns(config.MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(config.MySQL.ConnMaxLifetime)

	return db, nil
}
