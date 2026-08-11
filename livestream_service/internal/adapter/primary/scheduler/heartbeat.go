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

type Heartbeat struct {
	usecase  port.NodeSchedulerUsecase
	interval string
	cron     *cron.Cron
	cancel   context.CancelFunc
}

func NewHeartbeat(usecase port.NodeSchedulerUsecase, cfg *config.Config) *Heartbeat {
	return &Heartbeat{
		usecase:  usecase,
		interval: cfg.Transcode.HeartbeatIntervalDuration().String(),
	}
}

func (h *Heartbeat) Start(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	spec := fmt.Sprintf("@every %s", h.interval)
	if _, err := c.AddFunc(spec, func() {
		if err := h.usecase.Heartbeat(runCtx); err != nil {
			logger.WithContext(runCtx).Error("node heartbeat failed", zap.Error(err))
		}
	}); err != nil {
		cancel()
		return err
	}
	h.cron = c
	h.cron.Start()
	return nil
}

func (h *Heartbeat) Stop() {
	if h.cron == nil {
		return
	}
	<-h.cron.Stop().Done()
	if h.cancel != nil {
		h.cancel()
	}
}
