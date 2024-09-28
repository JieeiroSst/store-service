package consumer

import (
	"context"

	"github.com/JIeeiroSst/consumer-service/internal/usecase"
	"github.com/nats-io/nats.go"
)

type consumer struct {
	Usecase usecase.Usecase
	Nats    *nats.Conn
}

type ConsumerInterface interface {
	Start(ctx context.Context)
}

func NewConsumer(Usecase usecase.Usecase, Nats *nats.Conn) ConsumerInterface {
	return &consumer{
		Usecase: Usecase,
		Nats:    Nats,
	}
}

func (c *consumer) Start(ctx context.Context) {

}
