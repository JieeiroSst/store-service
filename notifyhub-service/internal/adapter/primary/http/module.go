package http

import (
	"github.com/JIeeiroSst/notifyhub-service/config"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewHandler),
	fx.Provide(newRateLimiter),
)

func newRateLimiter(lc fx.Lifecycle, cfg *config.Config) *RateLimiter {
	rl := NewRateLimiter(cfg.Server.RateLimit)
	lc.Append(fx.StopHook(rl.Close))
	return rl
}
