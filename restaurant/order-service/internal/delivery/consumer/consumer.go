package consumer

import (
	"github.com/JIeeiroSst/order-service/internal/usecase"
	"github.com/nats-io/nats.go"
)

type consumer struct {
	Usecase usecase.Usecase
	Nats    *nats.Conn
}

type ConsumerInterface interface {
	Start()
}

func NewConsumer(Usecase usecase.Usecase, Nats *nats.Conn) ConsumerInterface {
	return &consumer{
		Usecase: Usecase,
		Nats:    Nats,
	}
}

func (c *consumer) Start() {

}
