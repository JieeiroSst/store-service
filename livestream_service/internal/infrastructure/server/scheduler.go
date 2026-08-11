package server

import (
	"context"

	schedulerAdapter "github.com/JIeeiroSst/livestream-service/internal/adapter/primary/scheduler"
	"go.uber.org/fx"
)

type SchedulerParams struct {
	fx.In

	LC        fx.Lifecycle
	Heartbeat *schedulerAdapter.Heartbeat
	Watchdog  *schedulerAdapter.Watchdog
}

func NewSchedulerServer(p SchedulerParams) {
	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := p.Heartbeat.Start(ctx); err != nil {
				return err
			}
			return p.Watchdog.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			p.Heartbeat.Stop()
			p.Watchdog.Stop()
			return nil
		},
	})
}
