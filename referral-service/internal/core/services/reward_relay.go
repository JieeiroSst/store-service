package services

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/ports"
)

const (
	relayInterval = 30 * time.Second
	relayGrace    = 30 * time.Second
	relayBatch    = 100
)

var RelayModule = fx.Options(
	fx.Invoke(registerRewardRelay),
)

type rewardRelay struct {
	rewards   ports.RewardRepository
	publisher ports.RewardPublisher
	log       *zap.Logger
}

func registerRewardRelay(
	lc fx.Lifecycle,
	cfg *config.Config,
	rewards ports.RewardRepository,
	publisher ports.RewardPublisher,
	log *zap.Logger,
) {
	if !cfg.RabbitMQ.Enabled {
		return
	}
	rl := &rewardRelay{rewards: rewards, publisher: publisher, log: log.Named("reward-relay")}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				rl.run(ctx)
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			select {
			case <-done:
			case <-stopCtx.Done():
			}
			return nil
		},
	})
}

func (r *rewardRelay) run(ctx context.Context) {
	t := time.NewTicker(relayInterval)
	defer t.Stop()
	for {
		r.flush(ctx)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (r *rewardRelay) flush(ctx context.Context) {
	pending, err := r.rewards.FindUnpublished(ctx, time.Now().Add(-relayGrace).UnixMilli(), relayBatch)
	if err != nil {
		if ctx.Err() == nil {
			r.log.Error("load unpublished rewards", zap.Error(err))
		}
		return
	}
	for _, rw := range pending {
		if ctx.Err() != nil {
			return
		}
		if err := r.publisher.PublishRewardGranted(ctx, rewardEvent(rw)); err != nil {
			r.log.Warn("relay publish failed", zap.String("ref_code", rw.RefCode), zap.Error(err))
			return
		}
		if err := r.rewards.MarkPublished(ctx, rw.OwnerUserID, rw.RefCode, time.Now().UnixMilli()); err != nil {
			r.log.Warn("relay could not mark reward published", zap.String("ref_code", rw.RefCode), zap.Error(err))
		}
	}
	if len(pending) > 0 {
		r.log.Info("relayed unpublished rewards", zap.Int("count", len(pending)))
	}
}
