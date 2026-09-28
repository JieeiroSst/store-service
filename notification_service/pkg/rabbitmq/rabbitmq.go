package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/streadway/amqp"
)

type rabbitMQ struct {
	rabbitmq *amqp.Connection
	mu       sync.RWMutex
	closed   bool
	config   config.RabbitConfig

	pubMu      sync.Mutex
	pubCh      *amqp.Channel
	pubConfirm chan amqp.Confirmation

	delaysMu sync.RWMutex
	delays   map[string][]time.Duration
}

const (
	QueueNotifications = "notifications"
	QueueCampaignBatch = "notification.campaign.batch"

	headerAttempt  = "x-attempt"
	headerReason   = "x-failure-reason"
	confirmTimeout = 10 * time.Second
)

var ErrPublishNotConfirmed = errors.New("rabbitmq did not confirm the publish")

type Delivery struct {
	Body    []byte
	Attempt int
}

type Handler func(ctx context.Context, d Delivery) error

type RabbitMQ interface {
	PublishToQueue(notification *model.Notification) error
	DeclareQueue(queue string, retryDelays []time.Duration) error
	Publish(queue string, body []byte, attempt int) error
	Retry(queue string, body []byte, nextAttempt int) (scheduled bool, err error)
	DeadLetter(queue string, body []byte, reason string) error
	Consume(ctx context.Context, queue string, prefetch, workers int, handle Handler) error
}

var (
	instance *rabbitMQ
	once     sync.Once
)

func GetInstance(config config.RabbitConfig) (RabbitMQ, error) {
	once.Do(func() {
		instance = &rabbitMQ{
			config: config,
			delays: map[string][]time.Duration{},
		}
		if err := instance.connect(); err != nil {
			log.Printf("Failed to initialize RabbitMQ connection: %v", err)
		}

		go instance.monitorConnection()
	})

	return instance, nil
}

func (r *rabbitMQ) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.rabbitmq != nil && !r.rabbitmq.IsClosed() {
		return nil
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		r.config.Username,
		r.config.Password,
		r.config.Host,
		r.config.Port,
		r.config.VirtualHost,
	)

	var err error
	for i := 0; i < r.config.MaxRetries; i++ {
		r.rabbitmq, err = amqp.Dial(url)
		if err == nil {
			r.closed = false
			return nil
		}

		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v",
			i+1, r.config.MaxRetries, err)
		time.Sleep(r.config.RetryDelay)
	}

	return fmt.Errorf("failed to connect after %d attempts: %v",
		r.config.MaxRetries, err)
}

func (r *rabbitMQ) GetConnection() (*amqp.Connection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.rabbitmq == nil || r.rabbitmq.IsClosed() {
		return nil, fmt.Errorf("connection is not established")
	}

	return r.rabbitmq, nil
}

func (r *rabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.rabbitmq != nil && !r.rabbitmq.IsClosed() {
		r.closed = true
		return r.rabbitmq.Close()
	}

	return nil
}

func (r *rabbitMQ) monitorConnection() {
	for {
		if r.closed {
			return
		}

		r.mu.RLock()
		if r.rabbitmq == nil || r.rabbitmq.IsClosed() {
			r.mu.RUnlock()
			if err := r.connect(); err != nil {
				log.Printf("Failed to reconnect: %v", err)
				time.Sleep(r.config.RetryDelay)
				continue
			}
			log.Println("Successfully reconnected to RabbitMQ")
		} else {
			r.mu.RUnlock()
		}

		time.Sleep(time.Second * 5)
	}
}

func (r *rabbitMQ) CreateChannel() (*amqp.Channel, error) {
	rabbitmq, err := r.GetConnection()
	if err != nil {
		return nil, err
	}

	return rabbitmq.Channel()
}

func RetryQueue(queue string, tier int) string { return queue + ".retry." + strconv.Itoa(tier) }

func DeadLetterQueue(queue string) string { return queue + ".dlq" }

func (r *rabbitMQ) DeclareQueue(queue string, retryDelays []time.Duration) error {
	ch, err := r.CreateChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		return err
	}
	for i, d := range retryDelays {
		args := amqp.Table{
			"x-message-ttl":             int64(d / time.Millisecond),
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": queue,
		}
		if _, err := ch.QueueDeclare(RetryQueue(queue, i+1), true, false, false, false, args); err != nil {
			return err
		}
	}
	if _, err := ch.QueueDeclare(DeadLetterQueue(queue), true, false, false, false, nil); err != nil {
		return err
	}

	r.delaysMu.Lock()
	r.delays[queue] = retryDelays
	r.delaysMu.Unlock()
	return nil
}

func (r *rabbitMQ) PublishToQueue(notification *model.Notification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	return r.Publish(QueueNotifications, body, 1)
}

func (r *rabbitMQ) Publish(queue string, body []byte, attempt int) error {
	return r.publish(queue, body, amqp.Table{headerAttempt: int64(attempt)})
}

func (r *rabbitMQ) Retry(queue string, body []byte, nextAttempt int) (bool, error) {
	r.delaysMu.RLock()
	delays := r.delays[queue]
	r.delaysMu.RUnlock()
	tier := nextAttempt - 1
	if tier < 1 || tier > len(delays) {
		return false, nil
	}
	return true, r.publish(RetryQueue(queue, tier), body, amqp.Table{headerAttempt: int64(nextAttempt)})
}

func (r *rabbitMQ) DeadLetter(queue string, body []byte, reason string) error {
	if len(reason) > 1000 {
		reason = reason[:1000]
	}
	return r.publish(DeadLetterQueue(queue), body, amqp.Table{headerReason: reason})
}

func (r *rabbitMQ) publish(queue string, body []byte, headers amqp.Table) error {
	r.pubMu.Lock()
	defer r.pubMu.Unlock()

	if r.pubCh == nil {
		ch, err := r.CreateChannel()
		if err != nil {
			return err
		}
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			return err
		}
		r.pubCh = ch
		r.pubConfirm = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	}

	err := r.pubCh.Publish("", queue, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Headers:      headers,
		Body:         body,
	})
	if err != nil {
		r.resetPublisher()
		return err
	}

	select {
	case c, ok := <-r.pubConfirm:
		if !ok {
			r.resetPublisher()
			return ErrPublishNotConfirmed
		}
		if !c.Ack {
			return ErrPublishNotConfirmed
		}
		return nil
	case <-time.After(confirmTimeout):
		r.resetPublisher()
		return ErrPublishNotConfirmed
	}
}

func (r *rabbitMQ) resetPublisher() {
	if r.pubCh != nil {
		_ = r.pubCh.Close()
	}
	r.pubCh = nil
	r.pubConfirm = nil
}

func attemptOf(d amqp.Delivery) int {
	switch v := d.Headers[headerAttempt].(type) {
	case int64:
		return int(v)
	case int32:
		return int(v)
	case int:
		return v
	}
	return 1
}

func (r *rabbitMQ) Consume(ctx context.Context, queue string, prefetch, workers int, handle Handler) error {
	prefetch, workers = max(prefetch, 1), max(workers, 1)
	for ctx.Err() == nil {
		if err := r.consumeOnce(ctx, queue, prefetch, workers, handle); err != nil && ctx.Err() == nil {
			log.Printf("rabbitmq consumer %s stopped: %v, reconnecting", queue, err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
		}
	}
	return nil
}

func (r *rabbitMQ) consumeOnce(ctx context.Context, queue string, prefetch, workers int, handle Handler) error {
	ch, err := r.CreateChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}
	tag := fmt.Sprintf("%s-%d", queue, time.Now().UnixNano())
	msgs, err := ch.Consume(queue, tag, false, false, false, false, nil)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for d := range msgs {
				if err := handle(ctx, Delivery{Body: d.Body, Attempt: attemptOf(d)}); err != nil {
					log.Printf("rabbitmq %s: handler failed, requeueing: %v", queue, err)
					time.Sleep(time.Second)
					_ = d.Nack(false, true)
					continue
				}
				_ = d.Ack(false)
			}
		}()
	}

	closed := ch.NotifyClose(make(chan *amqp.Error, 1))
	select {
	case <-ctx.Done():
		_ = ch.Cancel(tag, false)
		wg.Wait()
		return nil
	case err := <-closed:
		wg.Wait()
		if err != nil {
			return err
		}
		return errors.New("channel closed")
	}
}
