package scheduler

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Watchdog struct {
	usecase  port.NodeSchedulerUsecase
	interval string
	cron     *cron.Cron
}

func NewWatchdog(usecase port.NodeSchedulerUsecase, cfg *config.Config) *Watchdog {
	return &Watchdog{
		usecase:  usecase,
		interval: cfg.Transcode.WatchdogCheckIntervalDuration().String(),
	}
}

func (w *Watchdog) Start(ctx context.Context) error {
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	spec := fmt.Sprintf("@every %s", w.interval)
	if _, err := c.AddFunc(spec, func() {
		if err := w.usecase.CheckStale(ctx); err != nil {
			logger.WithContext(ctx).Error("transcode watchdog check failed", zap.Error(err))
		}
	}); err != nil {
		return err
	}
	w.cron = c
	w.cron.Start()
	return nil
}

func (w *Watchdog) Stop() {
	if w.cron == nil {
		return
	}
	<-w.cron.Stop().Done()
}
