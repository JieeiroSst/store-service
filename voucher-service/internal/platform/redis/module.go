package redis

import "go.uber.org/fx"

var Module = fx.Module("redis-client", fx.Provide(NewClient))
