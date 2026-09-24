package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/fx"

	"github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/ports"
)

var Module = fx.Options(
	fx.Provide(NewRewardPublisher),
	fx.Invoke(registerLifecycle),
)

const publishTimeout = 5 * time.Second

type rewardPublisher struct {
	cfg config.RabbitMQConfig

	mu      sync.Mutex
	conn    *amqp.Connection
	ch      *amqp.Channel
	confirm chan amqp.Confirmation
}

func NewRewardPublisher(cfg *config.Config) ports.RewardPublisher {
	if !cfg.RabbitMQ.Enabled {
		return noopPublisher{}
	}
	return &rewardPublisher{cfg: cfg.RabbitMQ}
}

func registerLifecycle(lc fx.Lifecycle, p ports.RewardPublisher) {
	if rp, ok := p.(*rewardPublisher); ok {
		lc.Append(fx.Hook{OnStop: func(context.Context) error {
			rp.close()
			return nil
		}})
	}
}

func (p *rewardPublisher) PublishRewardGranted(ctx context.Context, ev ports.RewardGrantedEvent) error {
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ensureChannel(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	err = p.ch.PublishWithContext(ctx, p.cfg.Exchange, p.cfg.RoutingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    ev.EventID,
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err == nil {
		select {
		case c := <-p.confirm:
			if !c.Ack {
				err = fmt.Errorf("broker nacked reward event %s", ev.EventID)
			}
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	if err != nil {
		p.closeLocked()
		return err
	}
	return nil
}

func (p *rewardPublisher) ensureChannel() error {
	if p.ch != nil && !p.ch.IsClosed() {
		return nil
	}
	p.closeLocked()

	conn, err := amqp.DialConfig(p.cfg.URL(), amqp.Config{Dial: amqp.DefaultDial(publishTimeout)})
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}
	if err := ch.ExchangeDeclare(p.cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		_ = conn.Close()
		return fmt.Errorf("declare exchange: %w", err)
	}
	if err := ch.Confirm(false); err != nil {
		_ = conn.Close()
		return fmt.Errorf("enable confirms: %w", err)
	}
	p.conn, p.ch = conn, ch
	p.confirm = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	return nil
}

func (p *rewardPublisher) close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeLocked()
}

func (p *rewardPublisher) closeLocked() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn, p.ch, p.confirm = nil, nil, nil
}

type noopPublisher struct{}

func (noopPublisher) PublishRewardGranted(context.Context, ports.RewardGrantedEvent) error {
	return nil
}
