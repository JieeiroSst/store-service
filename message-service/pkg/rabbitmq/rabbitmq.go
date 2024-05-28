package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type QueueRabbitMQ struct {
	conn *amqp.Connection
}

func NewQueueRabbitMQ(dns string) QueueRabbitMQ {
	conn, err := amqp.Dial(dns)
	if err != nil {
		log.Fatalf("could not connect to rabbitmq: %v", err)
		panic(err)
	}
	return QueueRabbitMQ{conn: conn}
}

func (r *QueueRabbitMQ) Publish(q string, msg []byte) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	payload := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         msg,
	}

	if err := ch.Publish("", q, false, false, payload); err != nil {
		return fmt.Errorf("[Publish] failed to publish to queue %v", err)
	}

	return nil
}

func (r *QueueRabbitMQ) Subscribe(q string) (<-chan amqp.Delivery, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	return ch.Consume(q, "", false, false, false, false, nil)
}

func (r *QueueRabbitMQ) CreateQueueFanout(ctx context.Context, queues []string) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil
	}
	for _, v := range queues {
		if err := ch.ExchangeDeclare(
			v,        // name
			"fanout", // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *QueueRabbitMQ) CreateQueueDirect(ctx context.Context, queues []string) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil
	}
	for _, v := range queues {
		if err := ch.ExchangeDeclare(
			v,        // name
			"direct", // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *QueueRabbitMQ) CreateQueueTopic(ctx context.Context, queues []string) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil
	}
	for _, v := range queues {
		if err := ch.ExchangeDeclare(
			v,        // name
			"topic", // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		); err != nil {
			return err
		}
	}
	return nil
}
