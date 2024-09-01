package queue

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	pubsub "github.com/samber/go-amqp-pubsub"
	"github.com/samber/mo"
	"github.com/sirupsen/logrus"
)

type Queue struct {
	Conn *pubsub.Connection
}

type ProducerOptionsExchange struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
}

func NewQueue(rabbitmqURI string) *Queue {
	conn, err := pubsub.NewConnection("example-connection-1", pubsub.ConnectionOptions{
		URI: rabbitmqURI,
		Config: amqp.Config{
			Dial:      amqp.DefaultDial(time.Second),
			Heartbeat: 60 * time.Second,
		},
		LazyConnection: mo.Some(true),
	})
	if err != nil {
		panic(err)
	}

	return &Queue{
		Conn: conn,
	}
}

func (q *Queue) PublishMessages(data []byte, key string, exchange *ProducerOptionsExchange) {
	producerOptions := pubsub.ProducerOptionsExchange{
		Kind: mo.Some(pubsub.ExchangeKindTopic),
	}

	if exchange != nil {
		producerOptions.Name = mo.Some(exchange.Name)
	}
	producer := pubsub.NewProducer(q.Conn, "example-producer-1", pubsub.ProducerOptions{
		Exchange: producerOptions,
	})

	err := producer.Publish(key, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         data,
	})
	if err != nil {
		logrus.Error(err)
	}
	logrus.Infof("Published message [RK=%s] %s", key, string(data))
}

func (q *Queue) ConsumeMessages(key string, exchange []ProducerOptionsExchange) <-chan *amqp.Delivery {
	consumerOptions := pubsub.ConsumerOptions{
		Queue: pubsub.ConsumerOptionsQueue{
			Name: key,
		},
		Message: pubsub.ConsumerOptionsMessage{
			PrefetchCount: mo.Some(1000),
		},
		EnableDeadLetter: mo.Some(true),
	}
	var bindings []pubsub.ConsumerOptionsBinding

	if len(exchange) > 0 {
		for _, v := range exchange {
			bindings = append(bindings, pubsub.ConsumerOptionsBinding{
				ExchangeName: v.Name,
				RoutingKey:   v.Key,
			})
		}
	}
	consumerOptions.Bindings = bindings
	consumer := pubsub.NewConsumer(q.Conn, "example-consumer-1", consumerOptions)
	return consumer.Consume()
}
