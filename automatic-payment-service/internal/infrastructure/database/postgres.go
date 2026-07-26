package database

import (
	"github.com/JIeeiroSst/automatic-payment-service/config"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/utils/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	db := postgres.NewPostgresConn(postgres.PostgresConfig{
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
