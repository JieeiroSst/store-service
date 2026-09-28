package worker

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/scheduler"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func RunScheduler(
	lc fx.Lifecycle,
	sched *scheduler.Scheduler,
	jobs port.JobUsecase,
	templates port.TemplateUsecase,
	log *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			n, err := templates.WarmUp(ctx)
			if err != nil {
				log.Warn("template warm-up failed", zap.Error(err))
			} else {
				log.Info("templates pre-compiled", zap.Int("count", n))
			}
			if err := jobs.RestoreSchedules(ctx); err != nil {
				return fmt.Errorf("restore schedules: %w", err)
			}
			sched.Start()
			log.Info("scheduler started")
			return nil
		},
		OnStop: func(context.Context) error {
			return sched.Shutdown()
		},
	})
}
