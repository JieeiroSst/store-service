package consumer

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/delivery-service/internal/dto"
	"github.com/JIeeiroSst/delivery-service/internal/usecase"
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
	c.Nats.Subscribe("delivery.ship", func(msg *nats.Msg) {
		deliveryActive, err := c.Usecase.Deliveries.FindByActive(ctx)
		if err != nil {
			logger.ConfigZap().Error(err)
		}

		var addressDelivery dto.AddressDelivery
		if err := json.Unmarshal(msg.Data, &addressDelivery); err != nil {
			logger.ConfigZap().Error(err)
		}

		delivery := dto.BuildUpdate(deliveryActive, addressDelivery)
		if err := c.Usecase.Update(ctx, delivery.ShipID, *delivery); err != nil {
			logger.ConfigZap().Error(err)
		}
	})
}
