package jobs

import (
	"context"
	"log"
	"time"

	"github.com/JIeeiroSst/polymarket-service/config"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"go.uber.org/fx"
)

func RegisterRewardJob(lc fx.Lifecycle, rewards port.RewardUsecase, cfg *config.Config) {
	interval := cfg.Exchange.RewardEpochDuration()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				ticker := time.NewTicker(interval)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case now := <-ticker.C:
						if n, err := rewards.RunEpoch(ctx, now); err != nil {
							log.Printf("liquidity reward epoch failed: %v", err)
						} else if n > 0 {
							log.Printf("liquidity rewards paid to %d makers", n)
						}
					}
				}
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			select {
			case <-done:
			case <-stopCtx.Done():
			}
			return nil
		},
	})
}
