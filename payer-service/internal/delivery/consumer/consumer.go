package consumer

import (
	"context"

	"github.com/JIeeiroSst/payer-service/internal/usecase"
	"github.com/nats-io/nats.go"
)

type Consumer struct {
	usecase *usecase.Usecase
	nats    *nats.Conn
}

func NewConsumer(usecase *usecase.Usecase, nats *nats.Conn) *Consumer {
	return &Consumer{
		usecase: usecase,
		nats:    nats,
	}
}

func (c *Consumer) Start(ctx context.Context) {

}
