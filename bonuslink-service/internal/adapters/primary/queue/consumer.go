package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/bonuslink-service/internal/config"
	"github.com/JIeeiroSst/bonuslink-service/internal/core/domain"
	"github.com/JIeeiroSst/bonuslink-service/internal/core/ports"
)

var Module = fx.Options(
	fx.Invoke(RegisterConsumer),
)

type rewardMessage struct {
	EventID     string  `json:"event_id"`
	UserID      string  `json:"user_id"`
	RefCode     string  `json:"ref_code"`
	RewardType  string  `json:"reward_type"`
	RewardValue float64 `json:"reward_value"`
}

type Consumer struct {
	cfg config.RabbitMQConfig
	svc ports.BonusService
	log *zap.Logger
}

func RegisterConsumer(lc fx.Lifecycle, cfg *config.Config, svc ports.BonusService, log *zap.Logger) {
	if !cfg.RabbitMQ.Enabled {
		log.Info("rabbitmq consumer disabled")
		return
	}
	c := &Consumer{cfg: cfg.RabbitMQ, svc: svc, log: log.Named("reward-consumer")}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				c.run(ctx)
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

func (c *Consumer) run(ctx context.Context) {
	backoff := time.Second
	for ctx.Err() == nil {
		started := time.Now()
		err := c.consume(ctx)
		if ctx.Err() != nil {
			return
		}
		if time.Since(started) > time.Minute {
			backoff = time.Second
		}
		c.log.Error("consumer stopped, reconnecting", zap.Error(err), zap.Duration("in", backoff))
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (c *Consumer) consume(ctx context.Context) error {
	conn, err := amqp.Dial(c.cfg.URL())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}
	defer ch.Close()

	if err := c.declare(ch); err != nil {
		return err
	}

	deliveries, err := ch.Consume(c.cfg.Queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	connClosed := conn.NotifyClose(make(chan *amqp.Error, 1))
	c.log.Info("consuming rewards", zap.String("queue", c.cfg.Queue))

	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-connClosed:
			if e == nil {
				return errors.New("connection closed")
			}
			return e
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("delivery channel closed")
			}
			c.handle(ctx, d)
		}
	}
}

func (c *Consumer) declare(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(c.cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(c.cfg.Queue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	if err := ch.QueueBind(c.cfg.Queue, c.cfg.RoutingKey, c.cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}
	if err := ch.Qos(c.cfg.Prefetch, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}
	return nil
}

func (c *Consumer) handle(ctx context.Context, d amqp.Delivery) {
	var msg rewardMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		c.log.Error("dropping malformed message", zap.Error(err))
		_ = d.Reject(false)
		return
	}

	_, _, err := c.svc.RecordReward(ctx, ports.RecordRewardRequest{
		EventID:    msg.EventID,
		UserID:     msg.UserID,
		RefCode:    msg.RefCode,
		RewardType: domain.RewardType(msg.RewardType),
		Value:      msg.RewardValue,
	})
	switch {
	case err == nil:
		_ = d.Ack(false)
	case errors.Is(err, domain.ErrInvalidReward):
		c.log.Error("dropping invalid reward message", zap.Error(err), zap.String("event_id", msg.EventID))
		_ = d.Reject(false)
	default:
		c.log.Error("record reward failed, requeueing", zap.Error(err), zap.String("event_id", msg.EventID))
		_ = d.Nack(false, true)
	}
}
