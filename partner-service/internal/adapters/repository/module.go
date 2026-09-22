package repository

import (
	"fmt"

	"github.com/JIeeiroSst/partner-service/internal/adapters/cache"
	"github.com/JIeeiroSst/partner-service/internal/config"
	"github.com/JIeeiroSst/partner-service/internal/core/ports"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newGormDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Mysql.MysqlHost, cfg.Mysql.MysqlPort, cfg.Mysql.MysqlUser, cfg.Mysql.MysqlPassword, cfg.Mysql.MysqlDbname)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func newRedisCache(cfg *config.Config) (*cache.RedisCache, error) {
	return cache.NewRedisCache(cfg.Cache.Host, "")
}

var Module = fx.Options(
	fx.Provide(newGormDB),
	fx.Provide(newRedisCache),
	fx.Provide(fx.Annotate(
		NewDB,
		fx.As(new(ports.PartnerRepository)),
		fx.As(new(ports.PartnershipRepository)),
		fx.As(new(ports.PartnershipsPartnerRepository)),
		fx.As(new(ports.ProjectRepository)),
	)),
)
