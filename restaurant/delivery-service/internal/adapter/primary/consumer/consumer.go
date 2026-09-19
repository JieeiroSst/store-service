package consumer

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/delivery-service/internal/domain/model"
	"github.com/JIeeiroSst/delivery-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"github.com/nats-io/nats.go"
)

// shipRequest is the wire shape published on "delivery.ship".
type shipRequest struct {
	Name      string `json:"name" form:"name"`
	Address   string `json:"address" form:"address"`
	KitchenID int    `json:"kitchen_id"`
}

type Subscriber struct {
	usecase port.DeliveryUsecase
	nc      *nats.Conn
	subs    []*nats.Subscription
}

func NewSubscriber(usecase port.DeliveryUsecase, nc *nats.Conn) *Subscriber {
	return &Subscriber{usecase: usecase, nc: nc}
}

func (s *Subscriber) Start() error {
	sub, err := s.nc.Subscribe("delivery.ship", s.onDeliveryShip)
	if err != nil {
		return err
	}
	s.subs = append(s.subs, sub)
	return nil
}

func (s *Subscriber) Stop() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}
}

func (s *Subscriber) onDeliveryShip(msg *nats.Msg) {
	ctx := context.Background()

	active, err := s.usecase.FindByActive(ctx)
	if err != nil {
		logger.ConfigZap().Error(err)
		return
	}
	if active == nil {
		return
	}

	var req shipRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		logger.ConfigZap().Error(err)
		return
	}

	delivery := &model.Delivery{
		ShipID:    active.ShipID,
		Name:      req.Name,
		Address:   req.Address,
		KitchenID: req.KitchenID,
		Status:    model.StatusBusy,
	}

	if err := s.usecase.Update(ctx, delivery.ShipID, delivery); err != nil {
		logger.ConfigZap().Error(err)
	}
}
