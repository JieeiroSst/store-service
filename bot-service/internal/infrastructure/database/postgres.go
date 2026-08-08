package database

import (
	"github.com/JIeeiroSst/bot-service/config"
	"github.com/JIeeiroSst/utils/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) *gorm.DB {
	return postgres.NewPostgresConn(postgres.PostgresConfig{
		PostgresqlHost:     cfg.Postgres.PostgresqlHost,
		PostgresqlPort:     cfg.Postgres.PostgresqlPort,
		PostgresqlUser:     cfg.Postgres.PostgresqlUser,
		PostgresqlPassword: cfg.Postgres.PostgresqlPassword,
		PostgresqlDbname:   cfg.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  cfg.Postgres.PostgresqlSSLMode,
	})
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
)
