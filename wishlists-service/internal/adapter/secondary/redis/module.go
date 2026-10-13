package rediscache

import (
	"github.com/JIeeiroSst/wishlists-service/internal/config"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Decorate(decorateRepository, decorateUsers),
	fx.Provide(
		newRateLimiter,
		fx.Annotate(newHealth, fx.ResultTags(`group:"health,flatten"`)),
	),
)

func decorateRepository(base port.WishlistRepository, rdb *redis.Client, cfg *config.Config) port.WishlistRepository {
	if rdb == nil {
		return base
	}
	return NewRepository(base, rdb, cfg.Redis.Prefix, cfg.Redis.WishlistTTL, cfg.Redis.PopularTTL, cfg.Redis.OpTimeout)
}

func decorateUsers(base port.UserDirectory, rdb *redis.Client, cfg *config.Config) port.UserDirectory {
	if rdb == nil {
		return base
	}
	return NewUsers(base, rdb, cfg.Redis.Prefix, cfg.Redis.UserTTL, cfg.Redis.MissingTTL, cfg.Redis.OpTimeout)
}

func newRateLimiter(rdb *redis.Client, cfg *config.Config) port.RateLimiter {
	if rdb == nil {
		return NoopRateLimiter{}
	}
	return NewRateLimiter(rdb, cfg.Redis.Prefix, cfg.Redis.OpTimeout)
}

func newHealth(rdb *redis.Client) []port.HealthChecker {
	if rdb == nil {
		return nil
	}
	return []port.HealthChecker{Health{rdb: rdb}}
}
