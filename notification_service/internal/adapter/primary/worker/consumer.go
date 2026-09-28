package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/adapter/secondary/publisher"
	"github.com/JIeeiroSst/nofitifaction-service/internal/application"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/rabbitmq"
	"go.uber.org/fx"
)

const (
	plannerInterval    = time.Second
	sweepInterval      = 30 * time.Second
	staleTokenInterval = time.Hour
)

type Consumer struct {
	mq           rabbitmq.RabbitMQ
	notification port.NotificationUsecase
	campaigns    port.CampaignRunner
	devices      port.UserDeviceUsecase
	audit        port.AuditUsecase
	cfg          *config.Config
}

func NewConsumer(mq rabbitmq.RabbitMQ, notification port.NotificationUsecase, campaigns port.CampaignRunner, devices port.UserDeviceUsecase, audit port.AuditUsecase, cfg *config.Config) *Consumer {
	return &Consumer{mq: mq, notification: notification, campaigns: campaigns, devices: devices, audit: audit, cfg: cfg}
}

func (c *Consumer) purgeAudit(ctx context.Context) {
	if n, err := c.audit.PurgeExpired(ctx); err != nil {
		log.Printf("audit retention: %v", err)
	} else if n > 0 {
		log.Printf("audit retention: deleted %d deliveries", n)
	}
}

func (c *Consumer) cleanStaleTokens(ctx context.Context) {
	if n, err := c.devices.DeactivateStaleTokens(ctx); err != nil {
		log.Printf("stale token cleanup: %v", err)
	} else if n > 0 {
		log.Printf("stale token cleanup: deactivated %d tokens", n)
	}
}

func (c *Consumer) HandleNotification(ctx context.Context, d rabbitmq.Delivery) error {
	var notification model.Notification
	if err := json.Unmarshal(d.Body, &notification); err != nil {
		return c.mq.DeadLetter(rabbitmq.QueueNotifications, d.Body, "invalid json: "+err.Error())
	}
	maxAttempts := len(application.NotificationRetryDelays) + 1
	err := c.notification.Dispatch(ctx, &notification, d.Attempt, d.Attempt >= maxAttempts)
	switch {
	case err == nil, errors.Is(err, common.ErrPermanent):
		return nil
	case d.Attempt >= maxAttempts:
		return c.mq.DeadLetter(rabbitmq.QueueNotifications, d.Body, err.Error())
	}
	scheduled, pubErr := c.mq.Retry(rabbitmq.QueueNotifications, d.Body, d.Attempt+1)
	if pubErr != nil {
		return pubErr
	}
	if !scheduled {
		return c.mq.DeadLetter(rabbitmq.QueueNotifications, d.Body, err.Error())
	}
	return nil
}

func (c *Consumer) HandleBatch(ctx context.Context, d rabbitmq.Delivery) error {
	var msg publisher.BatchMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil || msg.BatchID == 0 {
		return c.mq.DeadLetter(rabbitmq.QueueCampaignBatch, d.Body, "invalid batch message")
	}
	outcome, err := c.campaigns.SendBatch(ctx, msg.BatchID)
	if errors.Is(err, common.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if outcome.Retry {
		if _, err := c.mq.Retry(rabbitmq.QueueCampaignBatch, d.Body, outcome.Attempt); err != nil {
			log.Printf("campaign batch %d: schedule retry failed, sweeper will republish: %v", msg.BatchID, err)
		}
	}
	return nil
}

func every(ctx context.Context, interval time.Duration, fn func(ctx context.Context)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn(ctx)
		}
	}
}

func (c *Consumer) plan(ctx context.Context) {
	for ctx.Err() == nil {
		worked, err := c.campaigns.PlanNext(ctx)
		if err != nil {
			log.Printf("campaign planner: %v", err)
			return
		}
		if !worked {
			return
		}
	}
}

func (c *Consumer) sweep(ctx context.Context) {
	if n, err := c.campaigns.RepublishStale(ctx); err != nil {
		log.Printf("campaign sweeper: %v", err)
	} else if n > 0 {
		log.Printf("campaign sweeper: republished %d batches", n)
	}
}

func positive(v, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

func registerLifecycle(lc fx.Lifecycle, c *Consumer) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := c.mq.DeclareQueue(rabbitmq.QueueNotifications, application.NotificationRetryDelays); err != nil {
				log.Printf("declare %s: %v", rabbitmq.QueueNotifications, err)
			}
			if err := c.mq.DeclareQueue(rabbitmq.QueueCampaignBatch, application.BatchRetryDelays); err != nil {
				log.Printf("declare %s: %v", rabbitmq.QueueCampaignBatch, err)
			}
			w, cp := c.cfg.Worker, c.cfg.Campaign
			run(func() {
				_ = c.mq.Consume(ctx, rabbitmq.QueueNotifications, positive(w.Prefetch, 32), positive(w.Concurrency, 8), c.HandleNotification)
			})
			run(func() {
				_ = c.mq.Consume(ctx, rabbitmq.QueueCampaignBatch, positive(cp.Prefetch, 8), positive(cp.Concurrency, 4), c.HandleBatch)
			})
			run(func() { every(ctx, plannerInterval, c.plan) })
			run(func() { every(ctx, sweepInterval, c.sweep) })
			run(func() {
				c.cleanStaleTokens(ctx)
				every(ctx, staleTokenInterval, c.cleanStaleTokens)
			})
			run(func() {
				c.purgeAudit(ctx)
				every(ctx, staleTokenInterval, c.purgeAudit)
			})
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			done := make(chan struct{})
			go func() { wg.Wait(); close(done) }()
			select {
			case <-done:
				return nil
			case <-stopCtx.Done():
				return stopCtx.Err()
			}
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewConsumer),
	fx.Invoke(registerLifecycle),
)
