package consumer

import (
	"context"

	"github.com/nats-io/nats.go"
)

type consumer struct {
	Nats *nats.Conn
}

type ConsumerInterface interface {
	Start(ctx context.Context)
}

func NewConsumer(Nats *nats.Conn) ConsumerInterface {
	return &consumer{

		Nats: Nats,
	}
}

func (c *consumer) Start(ctx context.Context) {

}
