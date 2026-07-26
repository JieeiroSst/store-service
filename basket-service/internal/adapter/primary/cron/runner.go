package cron

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

type Runner struct {
	c *cron.Cron
}

func NewRunner() *Runner {
	return &Runner{c: cron.New()}
}

func (r *Runner) registerJobs() error {
	if _, err := r.c.AddFunc("*/5 * * * *", func() {
		fmt.Println("Chạy job 1")
	}); err != nil {
		return err
	}
	if _, err := r.c.AddFunc("*/10 * * * *", func() {
		fmt.Println("Chạy job 2")
	}); err != nil {
		return err
	}
	return nil
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
