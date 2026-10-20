package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/fx"
)

const publishTimeout = 5 * time.Second

var Module = fx.Options(
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)

type Publisher struct {
	cfg config.RabbitMQConfig

	mu      sync.Mutex
	conn    *amqp.Connection
	ch      *amqp.Channel
	confirm chan amqp.Confirmation
}

func New(cfg *config.Config) port.RewardPublisher {
	if !cfg.RabbitMQ.Enabled {
		return disabled{}
	}
	return &Publisher{cfg: cfg.RabbitMQ}
}

func registerLifecycle(lc fx.Lifecycle, p port.RewardPublisher) {
	if rp, ok := p.(*Publisher); ok {
		lc.Append(fx.Hook{OnStop: func(context.Context) error {
			rp.mu.Lock()
			defer rp.mu.Unlock()
			rp.closeLocked()
			return nil
		}})
	}
}

func (p *Publisher) Enabled() bool { return true }

func (p *Publisher) PublishReward(ctx context.Context, ev model.Reward) error {
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.ensureChannel(); err != nil {
		return fmt.Errorf("%w: %v", model.ErrUnavailable, err)
	}
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()
	err = p.ch.PublishWithContext(ctx, p.cfg.Exchange, p.cfg.RoutingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    ev.EventID,
		Type:         "lucky.reward.granted",
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err == nil {
		select {
		case c := <-p.confirm:
			if !c.Ack {
				err = fmt.Errorf("broker nacked %s", ev.EventID)
			}
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	if err != nil {
		p.closeLocked()
		return fmt.Errorf("%w: publish %s: %v", model.ErrUnavailable, ev.EventID, err)
	}
	return nil
}

func (p *Publisher) ensureChannel() error {
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

func (p *Publisher) closeLocked() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn, p.ch, p.confirm = nil, nil, nil
}

type disabled struct{}

func (disabled) Enabled() bool { return false }

func (disabled) PublishReward(context.Context, model.Reward) error {
	return fmt.Errorf("%w: rabbitmq publisher is disabled", model.ErrUnavailable)
}
