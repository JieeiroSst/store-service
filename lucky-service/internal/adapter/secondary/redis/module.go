package rediscache

import (
	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Decorate(decorateCampaigns),
	fx.Provide(
		newRateLimiter,
		newWinners,
		newSignals,
		fx.Annotate(newHealth, fx.ResultTags(`group:"health,flatten"`)),
	),
)

func decorateCampaigns(base port.CampaignRepository, rdb *redis.Client, cfg *config.Config) port.CampaignRepository {
	if rdb == nil {
		return base
	}
	return NewCampaigns(base, rdb, cfg.Redis.Prefix, cfg.Redis.CampaignTTL, cfg.Redis.OpTimeout)
}

func newRateLimiter(rdb *redis.Client, cfg *config.Config) port.RateLimiter {
	if rdb == nil {
		return NoopRateLimiter{}
	}
	return NewRateLimiter(rdb, cfg.Redis.Prefix, cfg.Redis.OpTimeout)
}

func newWinners(rdb *redis.Client, cfg *config.Config) port.WinnerFeed {
	if rdb == nil {
		return nil
	}
	return NewWinners(rdb, cfg.Redis.Prefix, cfg.Redis.WinnersSize, cfg.Redis.OpTimeout)
}

func newSignals(rdb *redis.Client, cfg *config.Config) port.FraudSignals {
	if rdb == nil {
		return nil
	}
	return NewSignals(rdb, cfg.Redis.Prefix, cfg.Redis.OpTimeout)
}

func newHealth(rdb *redis.Client) []port.HealthChecker {
	if rdb == nil {
		return nil
	}
	return []port.HealthChecker{Health{rdb: rdb}}
}
