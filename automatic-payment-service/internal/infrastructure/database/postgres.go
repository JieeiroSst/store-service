package database

import (
	"fmt"
	"time"

	"github.com/JIeeiroSst/automatic-payment-service/config"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/utils/postgres"
	"go.uber.org/fx"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	db := openPostgres(postgres.PostgresConfig{
		PostgresqlHost:     cfg.Postgres.PostgresqlHost,
		PostgresqlPort:     cfg.Postgres.PostgresqlPort,
		PostgresqlUser:     cfg.Postgres.PostgresqlUser,
		PostgresqlPassword: cfg.Postgres.PostgresqlPassword,
		PostgresqlDbname:   cfg.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  cfg.Postgres.PostgresqlSSLMode,
	})

	if err := db.AutoMigrate(
		&model.Subscription{},
		&model.PaymentMethod{},
		&model.Transaction{},
		&model.Invoice{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
)

func openPostgres(c postgres.PostgresConfig) *gorm.DB {
	sslMode := "disable"
	if c.PostgresqlSSLMode {
		sslMode = "require"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.PostgresqlHost, c.PostgresqlUser, c.PostgresqlPassword, c.PostgresqlDbname, c.PostgresqlPort, sslMode)
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	return db
}
