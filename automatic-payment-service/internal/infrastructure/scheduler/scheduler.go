package scheduler

import (
	"context"
	"time"

	"github.com/JIeeiroSst/automatic-payment-service/config"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	LC           fx.Lifecycle
	Cfg          *config.Config
	Subscription port.SubscriptionUsecase
}

func New(p Params) {
	ticker := time.NewTicker(p.Cfg.Billing.RenewalCheckInterval)
	done := make(chan struct{})

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go run(ticker, done, p.Subscription)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			ticker.Stop()
			close(done)
			return nil
		},
	})
}

func run(ticker *time.Ticker, done chan struct{}, sub port.SubscriptionUsecase) {
	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			lg := logger.WithContext(ctx)

			count, err := sub.ProcessDueRenewals(ctx)
			if err != nil {
				lg.Error("scheduler: renewal batch failed", zap.Error(err))
				continue
			}
			lg.Info("scheduler: renewal batch complete", zap.Int("processed", count))
		case <-done:
			return
		}
	}
}
