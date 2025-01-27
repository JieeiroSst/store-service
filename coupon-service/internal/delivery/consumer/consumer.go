package consumer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type ConsumerConfig struct {
	Subject    string
	QueueGroup string
	NumWorkers int
	MaxRetries int
	RetryDelay time.Duration
}

type MessageHandler func(msg *nats.Msg) error

type MultiConsumer struct {
	nc         *nats.Conn
	consumers  []*Consumer
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type Consumer struct {
	config       ConsumerConfig
	handler      MessageHandler
	subscription *nats.Subscription
}

func NewMultiConsumer(url string) (*MultiConsumer, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MultiConsumer{
		nc:         nc,
		ctx:        ctx,
		cancelFunc: cancel,
	}, nil
}

func (mc *MultiConsumer) AddConsumer(config ConsumerConfig, handler MessageHandler) error {
	consumer := &Consumer{
		config:  config,
		handler: handler,
	}

	sub, err := mc.nc.QueueSubscribe(config.Subject, config.QueueGroup,
		func(msg *nats.Msg) {
			for attempt := 0; attempt < config.MaxRetries; attempt++ {
				err := handler(msg)
				if err == nil {
					return
				}

				log.Printf("Error processing message (attempt %d/%d): %v",
					attempt+1, config.MaxRetries, err)

				if attempt < config.MaxRetries-1 {
					time.Sleep(config.RetryDelay)
				}
			}
		})

	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %w",
			config.Subject, err)
	}

	consumer.subscription = sub
	mc.consumers = append(mc.consumers, consumer)
	return nil
}

func (mc *MultiConsumer) Start() {
	var wg sync.WaitGroup

	for _, consumer := range mc.consumers {
		for i := 0; i < consumer.config.NumWorkers; i++ {
			wg.Add(1)
			go func(workerID int, c *Consumer) {
				defer wg.Done()
				log.Printf("Started worker %d for subject %s",
					workerID, c.config.Subject)

				<-mc.ctx.Done()
				log.Printf("Stopping worker %d for subject %s",
					workerID, c.config.Subject)
			}(i, consumer)
		}
	}

	wg.Wait()
}

func (mc *MultiConsumer) Stop() error {
	mc.cancelFunc()

	for _, consumer := range mc.consumers {
		if err := consumer.subscription.Unsubscribe(); err != nil {
			return fmt.Errorf("failed to unsubscribe from %s: %w",
				consumer.config.Subject, err)
		}
	}

	return mc.nc.Drain()
}
