package outbox

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

const relayBatchSize = 50

type Relay struct {
	repo      Repository
	publisher EventPublisher
	interval  time.Duration
	log       *zap.Logger

	stop chan struct{}
	done chan struct{}
}

func NewRelay(repo Repository, publisher EventPublisher, interval time.Duration, log *zap.Logger) *Relay {
	return &Relay{
		repo:      repo,
		publisher: publisher,
		interval:  interval,
		log:       log,
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}
}

func (r *Relay) run() {
	defer close(r.done)
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-ticker.C:
			r.tick(context.Background())
		}
	}
}

func (r *Relay) tick(ctx context.Context) {
	events, err := r.repo.FetchUnpublished(ctx, relayBatchSize)
	if err != nil {
		r.log.Error("outbox: fetch unpublished failed", zap.Error(err))
		return
	}
	for _, evt := range events {
		if err := r.publisher.Publish(ctx, evt.Topic, evt.AggregateID, evt.Payload); err != nil {
			r.log.Warn("outbox: publish failed, will retry next tick",
				zap.String("event_id", evt.ID), zap.Error(err))
			_ = r.repo.MarkFailed(ctx, evt.ID, err.Error())
			continue
		}
		if err := r.repo.MarkPublished(ctx, evt.ID); err != nil {
			r.log.Error("outbox: mark published failed", zap.String("event_id", evt.ID), zap.Error(err))
		}
	}
}

func RegisterRelay(lc fx.Lifecycle, relay *Relay) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go relay.run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			close(relay.stop)
			select {
			case <-relay.done:
			case <-ctx.Done():
			}
			return nil
		},
	})
}
