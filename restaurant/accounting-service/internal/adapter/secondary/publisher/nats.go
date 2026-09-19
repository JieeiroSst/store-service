package publisher

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
	"github.com/JIeeiroSst/accounting-service/internal/domain/port"
	"github.com/nats-io/nats.go"
)


const deliveryShipSubject = "delivery.ship"

type natsPublisher struct {
	nc *nats.Conn
}

func NewNatsPublisher(nc *nats.Conn) port.OrderPublisher {
	return &natsPublisher{nc: nc}
}

func (p *natsPublisher) PublishOrderSuccess(ctx context.Context, order model.Order) error {
	return p.publish(model.OrderStatusSuccess, order)
}

func (p *natsPublisher) PublishOrderReject(ctx context.Context, order model.Order) error {
	return p.publish(model.OrderStatusReject, order)
}

func (p *natsPublisher) PublishDeliveryShip(ctx context.Context, delivery model.Delivery) error {
	return p.publish(deliveryShipSubject, delivery)
}

func (p *natsPublisher) publish(subject string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.nc.Publish(subject, data)
}
