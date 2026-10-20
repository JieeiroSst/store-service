package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"go.uber.org/fx"
)

type RewardRelay struct {
	outbox    port.RewardOutbox
	publisher port.RewardPublisher
	clock     port.Clock
	cfg       config.RelayConfig
}

func NewRewardRelay(cfg *config.Config, outbox port.RewardOutbox, publisher port.RewardPublisher, clock port.Clock) *RewardRelay {
	return &RewardRelay{outbox: outbox, publisher: publisher, clock: clock, cfg: cfg.Relay}
}

func (r *RewardRelay) Flush(ctx context.Context) (int, error) {
	pending, err := r.outbox.Pending(ctx, r.clock.Now().Add(-r.cfg.Grace), r.cfg.Batch)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, rw := range pending {
		if err := r.publisher.PublishReward(ctx, rw); err != nil {
			return sent, err
		}
		if err := r.outbox.MarkPublished(ctx, rw.ID, r.clock.Now()); err != nil {
			return sent, err
		}
		sent++
	}
	return sent, nil
}

func (r *RewardRelay) run(ctx context.Context) {
	t := time.NewTicker(r.cfg.Interval)
	defer t.Stop()
	for {
		n, err := r.Flush(ctx)
		switch {
		case err != nil && ctx.Err() == nil:
			slog.Warn("reward relay flush failed", "published", n, "err", err)
		case n > 0:
			slog.Info("reward relay published rewards", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func registerRelay(lc fx.Lifecycle, r *RewardRelay) {
	if !r.publisher.Enabled() {
		slog.Warn("RABBITMQ_ENABLED is false, won points stay queued in lucky_reward_outbox until the relay is enabled")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				r.run(ctx)
			}()
			return nil
		},
		OnStop: func(stop context.Context) error {
			cancel()
			select {
			case <-done:
			case <-stop.Done():
			}
			return nil
		},
	})
}
