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
	cancel   context.CancelFunc
}

func NewWatchdog(usecase port.NodeSchedulerUsecase, cfg *config.Config) *Watchdog {
	return &Watchdog{
		usecase:  usecase,
		interval: cfg.Transcode.WatchdogCheckIntervalDuration().String(),
	}
}

// Start ignores the fx lifecycle context beyond validating the initial call: that
// context is cancelled once app startup finishes, but the cron job below runs for
// the lifetime of the process, so it needs its own long-lived context tied to Stop.
func (w *Watchdog) Start(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	spec := fmt.Sprintf("@every %s", w.interval)
	if _, err := c.AddFunc(spec, func() {
		if err := w.usecase.CheckStale(runCtx); err != nil {
			logger.WithContext(runCtx).Error("transcode watchdog check failed", zap.Error(err))
		}
	}); err != nil {
		cancel()
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
	if w.cancel != nil {
		w.cancel()
	}
}
