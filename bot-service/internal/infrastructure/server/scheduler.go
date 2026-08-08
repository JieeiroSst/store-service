package server

import (
	"context"

	schedulerprimary "github.com/JIeeiroSst/bot-service/internal/adapter/primary/scheduler"
	"go.uber.org/fx"
)

type SchedulerParams struct {
	fx.In

	LC        fx.Lifecycle
	Scheduler *schedulerprimary.Scheduler
}

func NewSchedulerServer(p SchedulerParams) {
	ctx, cancel := context.WithCancel(context.Background())

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return p.Scheduler.Start(ctx)
		},
		OnStop: func(context.Context) error {
			cancel()
			p.Scheduler.Stop()
			return nil
		},
	})
}
