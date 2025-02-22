package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/dto"
	"github.com/streadway/amqp"
)

type rabbitMQ struct {
	rabbitmq *amqp.Connection
	mu       sync.RWMutex
	closed   bool
	config   config.RabbitConfig
}

type RabbitMQ interface {
	PublishToQueue(notification *dto.Notification) error
	StartConsumer(fn func(notification dto.Notification) error) error
}

var (
	instance *rabbitMQ
	once     sync.Once
)

func GetInstance(config config.RabbitConfig) (RabbitMQ, error) {
	once.Do(func() {
		instance = &rabbitMQ{
			config: config,
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

func (s *rabbitMQ) PublishToQueue(notification *dto.Notification) error {
	ch, err := s.rabbitmq.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"notifications",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

func (s *rabbitMQ) StartConsumer(fn func(notification dto.Notification) error) error {
	ch, err := s.rabbitmq.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"notifications",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var notification dto.Notification
			if err := json.Unmarshal(d.Body, &notification); err != nil {
				d.Nack(false, true)
				continue
			}
			if err := fn(notification); err != nil {
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
		}
	}()

	<-forever
	return nil
}
