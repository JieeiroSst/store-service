package redis

import (
	"github.com/JIeeiroSst/voucher-service/internal/platform/lock"
	"go.uber.org/fx"
)

var Module = fx.Module("redis-adapters",
	fx.Provide(
		fx.Annotate(NewLocker, fx.As(new(lock.Locker))),
		NewRateLimiter,
	),
)
