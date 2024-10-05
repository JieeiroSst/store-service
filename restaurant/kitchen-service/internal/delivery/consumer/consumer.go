package consumer

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/kitchen-service/internal/dto"
	"github.com/JIeeiroSst/kitchen-service/internal/usecase"
	"github.com/JieeiroSst/logger"
	"github.com/nats-io/nats.go"
)

type consumer struct {
	Usecase *usecase.Usecase
	Nats    *nats.Conn
}

type ConsumerInterface interface {
	Start(ctx context.Context)
}

func NewConsumer(Usecase *usecase.Usecase, Nats *nats.Conn) ConsumerInterface {
	return &consumer{
		Usecase: Usecase,
		Nats:    Nats,
	}
}

func (c *consumer) Start(ctx context.Context) {
	c.CreateKitchenSubscribe(ctx)
}

func (c *consumer) CreateKitchenSubscribe(ctx context.Context) {
	c.Nats.Subscribe("kitchen.create", func(msg *nats.Msg) {
		var customer dto.Customer
		if err := json.Unmarshal(msg.Data, &customer); err != nil {
			logger.ConfigZap().Error(err)
		}
		kitchen := customer.Build()
		if err := c.Usecase.Kitchens.Create(ctx, kitchen); err != nil {
			logger.ConfigZap().Error(err)
		}
	})
}
