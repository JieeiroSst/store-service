package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/redisstore"
	"github.com/JIeeiroSst/networking-service/internal/application"
	"github.com/JIeeiroSst/networking-service/internal/port"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type storeOut struct {
	fx.Out

	State      port.StateStore
	Catalog    port.CatalogStore
	KV         port.KVStore
	Sessions   port.SessionStore
	Intentions port.IntentionStore
	Leader     port.Leadership
}

func provideStore(lc fx.Lifecycle, cfg *config.Config) (storeOut, error) {
	switch cfg.Store.Backend {
	case "memory", "":
		s := memory.NewStore()
		return storeOut{State: s, Catalog: s, KV: s, Sessions: s, Intentions: s, Leader: application.SoloLeader{}}, nil

	case "redis":
		rdb := goredis.NewClient(&goredis.Options{
			Addr:     cfg.Store.RedisAddr,
			Password: cfg.Store.RedisPassword,
			DB:       cfg.Store.RedisDB,
		})
		s := redisstore.New(rdb, redisstore.Options{
			Prefix:       cfg.Store.RedisPrefix,
			PollInterval: cfg.Store.RedisPollInterval,
			OpTimeout:    cfg.Store.RedisOpTimeout,
		}, time.Now)
		elector := redisstore.NewElector(rdb, cfg.Store.RedisPrefix, cfg.Store.LeaderTTL)
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				if err := s.Start(ctx); err != nil {
					return err
				}
				elector.Start()
				return nil
			},
			OnStop: func(context.Context) error {
				elector.Stop()
				s.Close()
				return rdb.Close()
			},
		})
		return storeOut{State: s, Catalog: s, KV: s, Sessions: s, Intentions: s, Leader: elector}, nil
	}
	return storeOut{}, fmt.Errorf("unknown STORE_BACKEND %q (want memory or redis)", cfg.Store.Backend)
}
