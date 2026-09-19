package consumer

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JIeeiroSst/kitchen-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"github.com/nats-io/nats.go"
)

// customerMessage is the wire shape published on "kitchen.create" —
// distinct from model.Kitchen since it describes an incoming customer
// order rather than a kitchen record.
type customerMessage struct {
	Name      string          `json:"table_name"`
	KitchenID int             `json:"kitchen_id"`
	Menu      []menuFoodEntry `json:"menu"`
}

type menuFoodEntry struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Category model.Category `json:"category"`
}

func (c customerMessage) toKitchen() *model.Kitchen {
	foods := make([]model.Food, 0, len(c.Menu))
	for _, v := range c.Menu {
		foods = append(foods, model.Food{
			Name:     v.Name,
			Category: v.Category,
		})
	}
	return &model.Kitchen{
		Name:  c.Name,
		Foods: foods,
	}
}

type Subscriber struct {
	usecase port.KitchenUsecase
	nc      *nats.Conn
	subs    []*nats.Subscription
}

func NewSubscriber(usecase port.KitchenUsecase, nc *nats.Conn) *Subscriber {
	return &Subscriber{usecase: usecase, nc: nc}
}

func (s *Subscriber) Start() error {
	sub, err := s.nc.Subscribe("kitchen.create", s.onKitchenCreate)
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

func (s *Subscriber) onKitchenCreate(msg *nats.Msg) {
	var customer customerMessage
	if err := json.Unmarshal(msg.Data, &customer); err != nil {
		logger.ConfigZap().Error(err)
		return
	}

	if err := s.usecase.Create(context.Background(), customer.toKitchen()); err != nil {
		logger.ConfigZap().Error(err)
	}
}
