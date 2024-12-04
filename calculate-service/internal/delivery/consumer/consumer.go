package consumer

import (
	"github.com/JIeeiroSst/calculate-service/internal/usecase"
	"github.com/nats-io/nats.go"
)

type Consumer struct {
	nats    *nats.Conn
	usecase *usecase.Usecase
}

func NewConsumer(nats *nats.Conn, usecase *usecase.Usecase) *Consumer {
	return &Consumer{
		nats:    nats,
		usecase: usecase,
	}
}

func (c *Consumer) RunConsumer() {
	
}
