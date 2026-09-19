package publisher

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/consumer-service/internal/domain/model"
	"github.com/JIeeiroSst/consumer-service/internal/domain/port"
	"github.com/nats-io/nats.go"
)

type natsPublisher struct {
	nc *nats.Conn
}

func NewNatsPublisher(nc *nats.Conn) port.OrderPublisher {
	return &natsPublisher{nc: nc}
}

func (p *natsPublisher) PublishKitchenCreate(ctx context.Context, order model.PlaceOrder) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return p.nc.Publish("kitchen.create", data)
}

func (p *natsPublisher) PublishOrderCreated(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return p.nc.Publish("order.created", data)
}

func (p *natsPublisher) PublishCartAuthorization(ctx context.Context, cart model.CartAuthorization) error {
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return p.nc.Publish("accounting.authorize", data)
}
