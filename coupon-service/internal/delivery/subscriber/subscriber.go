package subscriber

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// SubscriptionType defines the type of NATS subscription
type SubscriptionType int

const (
	Simple SubscriptionType = iota
	Queue
	Wildcard
)

// SubscriptionConfig holds the configuration for each NATS subscription
type SubscriptionConfig struct {
	Subject     string
	QueueGroup  string
	Type        SubscriptionType
	DurableName string
	MaxInflight int
	AckWait     time.Duration
}

// MultiSubscriber manages multiple NATS subscriptions
type MultiSubscriber struct {
	nc            *nats.Conn
	js            nats.JetStreamContext
	subscriptions []*nats.Subscription
	handlers      map[string]nats.MsgHandler
	mu            sync.RWMutex
	ctx           context.Context
	cancelFunc    context.CancelFunc
}

// NewMultiSubscriber creates a new instance of MultiSubscriber
func NewMultiSubscriber(url string) (*MultiSubscriber, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MultiSubscriber{
		nc:         nc,
		js:         js,
		handlers:   make(map[string]nats.MsgHandler),
		ctx:        ctx,
		cancelFunc: cancel,
	}, nil
}

// Subscribe adds a new subscription with the specified configuration and handler
func (ms *MultiSubscriber) Subscribe(config SubscriptionConfig, handler nats.MsgHandler) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var sub *nats.Subscription
	var err error

	switch config.Type {
	case Simple:
		sub, err = ms.nc.Subscribe(config.Subject, handler)

	case Queue:
		sub, err = ms.nc.QueueSubscribe(config.Subject, config.QueueGroup, handler)

	case Wildcard:
		sub, err = ms.nc.Subscribe(config.Subject, handler)

	default:
		return fmt.Errorf("unknown subscription type: %v", config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create subscription for %s: %w",
			config.Subject, err)
	}

	ms.subscriptions = append(ms.subscriptions, sub)
	ms.handlers[config.Subject] = handler

	log.Printf("Successfully subscribed to subject: %s", config.Subject)
	return nil
}

// SubscribeJetStream adds a new JetStream subscription
func (ms *MultiSubscriber) SubscribeJetStream(config SubscriptionConfig, handler nats.MsgHandler) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var opts []nats.SubOpt

	if config.MaxInflight > 0 {
		opts = append(opts, nats.MaxAckPending(config.MaxInflight))
	}

	if config.AckWait > 0 {
		opts = append(opts, nats.AckWait(config.AckWait))
	}

	if config.DurableName != "" {
		opts = append(opts, nats.Durable(config.DurableName))
	}

	var sub *nats.Subscription
	var err error

	if config.QueueGroup != "" {
		sub, err = ms.js.QueueSubscribe(config.Subject, config.QueueGroup, handler, opts...)
	} else {
		sub, err = ms.js.Subscribe(config.Subject, handler, opts...)
	}

	if err != nil {
		return fmt.Errorf("failed to create JetStream subscription for %s: %w",
			config.Subject, err)
	}

	ms.subscriptions = append(ms.subscriptions, sub)
	ms.handlers[config.Subject] = handler

	log.Printf("Successfully created JetStream subscription for subject: %s",
		config.Subject)
	return nil
}

// Unsubscribe removes a subscription for the specified subject
func (ms *MultiSubscriber) Unsubscribe(subject string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for i, sub := range ms.subscriptions {
		if sub.Subject == subject {
			if err := sub.Unsubscribe(); err != nil {
				return fmt.Errorf("failed to unsubscribe from %s: %w",
					subject, err)
			}

			ms.subscriptions = append(ms.subscriptions[:i],
				ms.subscriptions[i+1:]...)
			delete(ms.handlers, subject)

			log.Printf("Successfully unsubscribed from subject: %s", subject)
			return nil
		}
	}

	return fmt.Errorf("subscription not found for subject: %s", subject)
}

// Stop gracefully closes all subscriptions
func (ms *MultiSubscriber) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.cancelFunc()

	for _, sub := range ms.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
	}

	return ms.nc.Drain()
}
