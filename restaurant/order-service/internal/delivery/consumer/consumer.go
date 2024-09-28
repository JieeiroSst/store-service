package consumer

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/order-service/internal/dto"
	"github.com/JIeeiroSst/order-service/internal/usecase"
	"github.com/JieeiroSst/logger"
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
	c.CreateOrderSubscribe(ctx)
	c.RejectOrderSubscribe(ctx)
	c.ReceivedItemSubscribe(ctx)
}

func (c *consumer) CreateOrderSubscribe(ctx context.Context) {
	c.Nats.Subscribe("order.created", func(msg *nats.Msg) {
		var order dto.Order

		if err := json.Unmarshal(msg.Data, &order); err != nil {
			logger.ConfigZap().Errorf("%v", err)
		}

		if err := c.Usecase.Orders.CreateOrder(ctx, order); err != nil {
			logger.ConfigZap().Errorf("%v", err)
		}
	})
}

func (c *consumer) RejectOrderSubscribe(ctx context.Context) {
	c.Nats.Subscribe("order.reject", func(msg *nats.Msg) {
		var order dto.Order

		if err := json.Unmarshal(msg.Data, &order); err != nil {
			logger.ConfigZap().Errorf("%v", err)
		}

		if err := c.Usecase.CancelOrder(ctx, order.ID, order); err != nil {
			logger.ConfigZap().Errorf("%v", err)
		}
	})
}

func (c *consumer) ReceivedItemSubscribe(ctx context.Context) {
	c.Nats.Subscribe("order.success", func(msg *nats.Msg) {
		var order dto.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			logger.ConfigZap().Errorf("%v", err)
		}

		if err := c.Usecase.SuccessOrder(ctx, order.ID, order); err != nil {
			logger.ConfigZap().Errorf("%v", err)
		}
	})
}
