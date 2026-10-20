package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/JIeeiroSst/point-service/config"
	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/domain/service"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Outcome int

const (
	Ack Outcome = iota
	Drop
	Retry
)

type RewardConsumer struct {
	cfg     config.RabbitMQConfig
	ledger  service.PointLedgerService
	backoff time.Duration
}

func NewRewardConsumer(cfg config.RabbitMQConfig, ledger service.PointLedgerService) *RewardConsumer {
	return &RewardConsumer{cfg: cfg, ledger: ledger, backoff: time.Second}
}

func (c *RewardConsumer) Handle(ctx context.Context, body []byte) Outcome {
	var ev dto.RewardEventDTO
	if err := json.Unmarshal(body, &ev); err != nil {
		log.Printf("reward consumer: drop undecodable message: %v", err)
		return Drop
	}
	applied, err := c.ledger.Apply(ctx, ev)
	switch {
	case errors.Is(err, service.ErrInvalidRewardEvent):
		log.Printf("reward consumer: drop %s: %v", ev.EventID, err)
		return Drop
	case err != nil:
		log.Printf("reward consumer: retry %s: %v", ev.EventID, err)
		return Retry
	case !applied:
		log.Printf("reward consumer: %s already credited, skipping", ev.EventID)
	}
	return Ack
}

func (c *RewardConsumer) url() string {
	u := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(c.cfg.User, c.cfg.Password),
		Host:   net.JoinHostPort(c.cfg.Host, c.cfg.Port),
		Path:   "/" + strings.TrimPrefix(c.cfg.VHost, "/"),
	}
	return u.String()
}

func (c *RewardConsumer) Run(ctx context.Context) {
	wait := c.backoff
	for ctx.Err() == nil {
		err := c.consume(ctx)
		if ctx.Err() != nil {
			return
		}
		log.Printf("reward consumer: %v, reconnecting in %s", err, wait)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		if wait *= 2; wait > 30*time.Second {
			wait = 30 * time.Second
		}
	}
}

func (c *RewardConsumer) consume(ctx context.Context) error {
	conn, err := amqp.DialConfig(c.url(), amqp.Config{Dial: amqp.DefaultDial(5 * time.Second)})
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}
	if err := ch.Qos(c.cfg.Prefetch, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}
	if _, err := ch.QueueDeclare(c.cfg.Queue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	for _, b := range c.cfg.Bindings {
		exchange, key, ok := strings.Cut(b, ":")
		if !ok || exchange == "" || key == "" {
			return fmt.Errorf("binding %q must be exchange:routing_key", b)
		}
		if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare exchange %s: %w", exchange, err)
		}
		if err := ch.QueueBind(c.cfg.Queue, key, exchange, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", b, err)
		}
	}
	deliveries, err := ch.Consume(c.cfg.Queue, "point-service", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	log.Printf("reward consumer: listening on %s", c.cfg.Queue)
	closed := conn.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-closed:
			return fmt.Errorf("connection closed: %v", e)
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("delivery channel closed")
			}
			switch c.Handle(ctx, d.Body) {
			case Ack:
				err = d.Ack(false)
			case Drop:
				err = d.Nack(false, false)
			case Retry:
				select {
				case <-ctx.Done():
				case <-time.After(c.backoff):
				}
				err = d.Nack(false, true)
			}
			if err != nil {
				return fmt.Errorf("settle delivery: %w", err)
			}
		}
	}
}
