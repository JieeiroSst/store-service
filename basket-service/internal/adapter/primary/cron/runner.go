package cron

import (
	"context"
	"log"
	"time"

	"github.com/JIeeiroSst/basket-service/internal/domain/port"

	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

type Runner struct {
	c        *cron.Cron
	checkout port.CheckoutUsecase
}

func NewRunner(checkout port.CheckoutUsecase) *Runner {
	return &Runner{c: cron.New(), checkout: checkout}
}

func (r *Runner) registerJobs() error {
	_, err := r.c.AddFunc("* * * * *", r.retryFulfillment)
	return err
}

func (r *Runner) retryFulfillment() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	n, err := r.checkout.RetryFulfillment(ctx, 50)
	if err != nil {
		log.Printf("cron: retry order fulfillment: %v", err)
		return
	}
	if n > 0 {
		log.Printf("cron: started %d order workflows for paid baskets", n)
	}
}

func registerLifecycle(lc fx.Lifecycle, r *Runner) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := r.registerJobs(); err != nil {
				return err
			}
			r.c.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopCtx := r.c.Stop()
			select {
			case <-stopCtx.Done():
			case <-ctx.Done():
			}
			return nil
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewRunner),
	fx.Invoke(registerLifecycle),
)
