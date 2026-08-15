package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

type Scheduler struct {
	cron *cron.Cron
}

func New() *Scheduler {
	return &Scheduler{cron: cron.New()}
}

func (s *Scheduler) AddFunc(spec string, cmd func()) error {
	_, err := s.cron.AddFunc(spec, cmd)
	return err
}

func RegisterLifecycle(lc fx.Lifecycle, s *Scheduler) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.cron.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			done := s.cron.Stop()
			select {
			case <-done.Done():
			case <-ctx.Done():
			}
			return nil
		},
	})
}
