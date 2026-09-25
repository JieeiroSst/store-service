package jobs

import (
	"context"
	"log"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
	"go.uber.org/fx"
)

func New(lc fx.Lifecycle, m port.ModelMaintainer, cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				run(ctx, m, cfg)
			}()
			return nil
		},
		OnStop: func(stop context.Context) error {
			cancel()
			select {
			case <-done:
			case <-stop.Done():
			}
			return nil
		},
	})
}

func run(ctx context.Context, m port.ModelMaintainer, cfg *config.Config) {
	sync := func() {
		if cfg.VideoService.BaseURL == "" {
			if err := m.Train(ctx); err != nil {
				log.Printf("train: %v", err)
			}
			return
		}
		if err := m.Sync(ctx); err != nil {
			log.Printf("sync: %v", err)
		}
	}
	sync()

	syncTick := time.NewTicker(cfg.Train.SyncInterval)
	trainTick := time.NewTicker(cfg.Train.TrainInterval)
	defer syncTick.Stop()
	defer trainTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-syncTick.C:
			sync()
		case <-trainTick.C:
			if err := m.Train(ctx); err != nil {
				log.Printf("train: %v", err)
			}
			if err := m.Cleanup(ctx); err != nil {
				log.Printf("cleanup snapshots: %v", err)
			}
		}
	}
}
