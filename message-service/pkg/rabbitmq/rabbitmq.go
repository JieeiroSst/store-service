package rabbitmq

import (
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
