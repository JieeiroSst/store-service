package buffer

import (
	"context"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Decorate(func(s port.VideoStorage, cfg *config.Config) port.VideoStorage {
		return NewChunkStorage(s, cfg)
	}),
	fx.Decorate(func(lc fx.Lifecycle, r port.VideoRepository, bus port.InvalidationBus, cfg *config.Config) port.VideoRepository {
		repo := NewCachedRepository(r, bus, cfg)
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error { go repo.Listen(ctx); return nil },
			OnStop:  func(context.Context) error { cancel(); return nil },
		})
		return repo
	}),
)
