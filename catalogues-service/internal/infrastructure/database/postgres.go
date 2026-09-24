package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/JIeeiroSst/catalogues-service/config"
	"github.com/JIeeiroSst/catalogues-service/internal/adapter/secondary/repository"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const migrationLockID = 727_002

func dsn(c config.PostgresConfig, db string) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, c.Port),
		Path:     "/" + db,
		RawQuery: url.Values{"sslmode": {c.SSLMode}}.Encode(),
	}
	return u.String()
}

func New(cfg *config.Config) (*gorm.DB, error) {
	c := cfg.Postgres

	if err := ensureDatabase(dsn(c, "postgres"), c.DBName); err != nil {
		return nil, fmt.Errorf("ensure database: %w", err)
	}

	db, err := gorm.Open(postgres.Open(dsn(c, c.DBName)), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	err = db.Connection(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_lock(?)", migrationLockID).Error; err != nil {
			return err
		}
		defer tx.Exec("SELECT pg_advisory_unlock(?)", migrationLockID)
		return tx.AutoMigrate(repository.Entities()...)
	})
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func ensureDatabase(adminDSN, name string) error {
	conn, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var exists bool
	if err := conn.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = conn.ExecContext(ctx, `CREATE DATABASE "`+name+`"`)
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
