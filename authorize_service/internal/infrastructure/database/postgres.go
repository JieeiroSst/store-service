package database

import (
	"github.com/JIeeiroSst/utils/postgres"
	"github.com/JieeiroSst/authorize-service/config"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	casbinpersist "github.com/casbin/casbin/v2/persist"
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

func NewCasbinAdapter(db *gorm.DB) (casbinpersist.Adapter, error) {
	return gormadapter.NewAdapterByDB(db)
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
	fx.Provide(NewCasbinAdapter),
)
