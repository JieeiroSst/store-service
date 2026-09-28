package worker

import (
	"context"
	"log"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/config"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(StartExpirySweeper, StartAlertDispatcher, StartPaymentReconciler),
)

func StartExpirySweeper(lc fx.Lifecycle, vending port.VendingService, cfg *config.Config) {
	every(lc, cfg.Vending.SweepInterval, func(ctx context.Context) {
		r, s, err := vending.ExpireStale(ctx)
		if err != nil && ctx.Err() == nil {
			log.Printf("expiry sweep: %v", err)
		}
		if r > 0 || s > 0 {
			log.Printf("expiry sweep: released %d reservations, %d sessions", r, s)
		}
	})
}

func StartAlertDispatcher(lc fx.Lifecycle, alerts port.AlertService, cfg *config.Config) {
	every(lc, cfg.Vending.AlertInterval, func(ctx context.Context) {
		for {
			sent, err := alerts.DispatchPending(ctx)
			if err != nil {
				if ctx.Err() == nil {
					log.Printf("alert dispatch: %v", err)
				}
				return
			}
			if sent == 0 {
				return
			}
			log.Printf("alert dispatch: sent %d alerts", sent)
		}
	})
}

func StartPaymentReconciler(lc fx.Lifecycle, vending port.VendingService, cfg *config.Config) {
	every(lc, cfg.Vending.ReconcileInterval, func(ctx context.Context) {
		settled, err := vending.ReconcilePayments(ctx)
		if err != nil && ctx.Err() == nil {
			log.Printf("payment reconcile: %v", err)
		}
		if settled > 0 {
			log.Printf("payment reconcile: settled %d pending payments", settled)
		}
	})
}

func every(lc fx.Lifecycle, interval time.Duration, run func(ctx context.Context)) {
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
					case <-ticker.C:
						run(ctx)
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
