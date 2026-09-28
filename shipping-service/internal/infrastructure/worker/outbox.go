package worker

import (
	"context"
	"log"
	"time"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"go.uber.org/fx"
)

func RunOutbox(lc fx.Lifecycle, outbox port.OutboxUsecase, cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	interval := cfg.Shipping.OutboxIntervalDuration()

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
					case <-ticker.C:
						for {
							n, err := outbox.DispatchDue(ctx)
							if err != nil {
								if ctx.Err() == nil {
									log.Printf("outbox dispatch: %v", err)
								}
								break
							}
							if n == 0 {
								break
							}
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
				return nil
			case <-stopCtx.Done():
				return stopCtx.Err()
			}
		},
	})
}
