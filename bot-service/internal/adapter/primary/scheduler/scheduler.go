package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/bot-service/config"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const defaultPollInterval = 30 * time.Second

type Scheduler struct {
	usecase      port.ContentUsecase
	pollInterval time.Duration
	cron         *cron.Cron
}

func NewScheduler(usecase port.ContentUsecase, cfg *config.Config) *Scheduler {
	interval, err := time.ParseDuration(cfg.Scheduler.PollInterval)
	if err != nil || interval <= 0 {
		interval = defaultPollInterval
	}
	return &Scheduler{usecase: usecase, pollInterval: interval}
}

func (s *Scheduler) Start(ctx context.Context) error {
	lg := logger.WithContext(ctx)

	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	spec := fmt.Sprintf("@every %s", s.pollInterval)
	if _, err := c.AddFunc(spec, func() {
		if err := s.usecase.PublishDuePosts(ctx); err != nil {
			lg.Error("scheduler: publish due posts", zap.Error(err))
		}
	}); err != nil {
		return err
	}

	s.cron = c
	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop() {
	if s.cron == nil {
		return
	}
	<-s.cron.Stop().Done()
}
